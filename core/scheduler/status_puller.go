package scheduler

import (
	"context"
	"time"

	"chihqiang/msg-push/core/sender"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
)

// StatusPuller 短信状态主动查询扫描器。
// 短信已发送（sending）且等待回调窗口已过仍无回执时，主动调用服务商 QueryStatus API
// 查询真实状态并更新任务终态，作为回调的主动补单手段。
// 与 SMSTimeoutScanner 互补：本扫描器"问服务商拿真实状态"，超时扫描器"直接判失败"兜底。
//
// 多实例安全：间隔级 Redis 分布式锁保证同一时刻仅一个实例执行；
// 条件化 UPDATE（CAS）保证不覆盖并发到达的真实回调；内存去重避免短窗口内对同一任务
// 重复调用服务商查询 API。
type StatusPuller struct {
	svc        *svc.ServiceContext
	interval   time.Duration
	queryAfter time.Duration // 发送后多久无回执开始主动查询
	dedupFor   time.Duration // 同一任务去重查询窗口
	limit      int
	lockKey    string
	recent     map[string]time.Time // 内存去重：task_id -> 上次查询时间
	stopCh     chan struct{}
	doneCh     chan struct{}
}

// NewStatusPuller 创建状态查询扫描器。
func NewStatusPuller(s *svc.ServiceContext) *StatusPuller {
	return &StatusPuller{
		svc:        s,
		interval:   60 * time.Second,
		queryAfter: 60 * time.Second,
		dedupFor:   5 * time.Minute,
		limit:      100,
		lockKey:    "msgpush:lock:status_puller",
		recent:     make(map[string]time.Time),
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
	}
}

// Start 启动扫描器（后台协程）。
func (s *StatusPuller) Start() {
	go s.run()
}

// Stop 优雅停止。
func (s *StatusPuller) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

// StartService 适配 service.Starter。
func (s *StatusPuller) StartService() { s.Start() }

// run 主循环。
func (s *StatusPuller) run() {
	defer close(s.doneCh)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.scanOnce()
		case <-s.stopCh:
			return
		}
	}
}

// scanOnce 执行一次扫描（带分布式锁）。
func (s *StatusPuller) scanOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 分布式锁：TTL 覆盖扫描耗时
	ok, err := s.svc.Redis.Client().SetNX(ctx, s.lockKey, 1, 2*s.interval).Result()
	if err != nil || !ok {
		logger.Infof("status puller skipped: lock unavailable")
		return
	}
	defer s.svc.Redis.Client().Del(context.Background(), s.lockKey)

	s.pullOnce(ctx)
}

// pullOnce 查询候选任务并逐个主动查询状态。
func (s *StatusPuller) pullOnce(ctx context.Context) {
	now := time.Now()
	cutoff := now.Add(-s.queryAfter)

	var tasks []model.PushTask
	// 覆盖无回执(空)与已标记超时(timeout)的 sending 任务：
	// sms_timeout_scanner 会在发送后置 callback_status=timeout，若不覆盖则这些任务
	// 永远不会被主动查询，只能等硬超时转 failed——即使服务商实际已送达。
	// 这里在硬超时前最后一次确认真实状态，避免误判。
	if err := s.svc.DB.WithContext(ctx).
		Where("message_type = ? AND status = ? AND (callback_status IS NULL OR callback_status = '' OR callback_status = ?)",
			string(model.ChannelTypeSMS), string(model.PushTaskStatusSending), "timeout").
		Where("(sent_at IS NOT NULL AND sent_at < ?) OR (sent_at IS NULL AND updated_at < ?)", cutoff, cutoff).
		Limit(s.limit).
		Find(&tasks).Error; err != nil {
		logger.Warnf("status puller: list tasks failed: %v", err)
		return
	}

	handled := 0
	for i := range tasks {
		t := &tasks[i]
		if s.deduped(t.TaskID, now) {
			continue
		}
		if s.queryAndApply(ctx, t, now) {
			handled++
		}
	}
	if len(tasks) > 0 {
		logger.Infof("status puller: scanned %d tasks, handled %d", len(tasks), handled)
	}

	// 清理过期去重记录
	s.cleanRecent(now)
}

// deduped 判断任务在去重窗口内是否已查询过。
func (s *StatusPuller) deduped(taskID string, now time.Time) bool {
	last, ok := s.recent[taskID]
	return ok && now.Sub(last) < s.dedupFor
}

// cleanRecent 清理过期去重记录。
func (s *StatusPuller) cleanRecent(now time.Time) {
	for k, v := range s.recent {
		if now.Sub(v) >= s.dedupFor {
			delete(s.recent, k)
		}
	}
}

// queryAndApply 查询单个任务状态并应用；返回是否处理（更新了状态或明确无状态可等）。
func (s *StatusPuller) queryAndApply(ctx context.Context, task *model.PushTask, now time.Time) bool {
	// 1. 取最新一条成功日志拿 provider_msg_id
	var pushLog model.PushLog
	if err := s.svc.DB.WithContext(ctx).
		Where("task_no = ? AND provider_msg_id <> ''", task.TaskID).
		Order("id DESC").First(&pushLog).Error; err != nil {
		logger.Debugf("status puller: no provider msg id for task %s", task.TaskID)
		return false
	}

	// 2. 加载服务商账号
	var pa model.ProviderAccount
	if err := s.svc.DB.WithContext(ctx).Where("id = ?", task.ProviderAccountID).First(&pa).Error; err != nil {
		logger.Debugf("status puller: provider account not found task=%s", task.TaskID)
		return false
	}

	// 3. 获取发送器并断言 StatusQuerier
	sd, err := sender.DefaultResolver.GetSender(pa.ProviderCode)
	if err != nil {
		return false
	}
	querier, ok := sd.(sender.StatusQuerier)
	if !ok {
		return false // 该服务商不支持主动查询
	}

	sendDate := now
	if task.SentAt != nil {
		sendDate = *task.SentAt
	}
	res, err := querier.QueryStatus(ctx, &sender.StatusQueryRequest{
		Task:            task,
		ProviderAccount: &pa,
		ProviderMsgID:   pushLog.ProviderMsgID,
		PhoneNumber:     task.Receiver,
		SendDate:        sendDate,
	})
	if err != nil {
		logger.Warnf("status puller: query task %s failed: %v", task.TaskID, err)
		return false
	}

	// 4. 应用查询结果（取第一个明确状态）
	// 批量场景下同一 BizId 可能返回多条记录，仅应用手机号匹配本任务的记录，
	// 服务商未返回手机号时放行（避免漏掉单条查询场景）。
	queried := false
	for _, r := range res.Results {
		if r.PhoneNumber != "" && r.PhoneNumber != task.Receiver {
			continue // 非本任务记录，跳过
		}
		switch r.Status {
		case "delivered":
			s.applyResult(ctx, task, &pa, &pushLog, r, now, true)
			queried = true
		case "failed":
			s.applyResult(ctx, task, &pa, &pushLog, r, now, false)
			queried = true
		}
		if queried {
			break
		}
	}
	if !queried {
		logger.Debugf("status puller: task %s query returned no definitive status", task.TaskID)
	}
	return queried
}

// applyResult 应用查询结果：更新任务终态（CAS）+ 更新 push_log + 触发 webhook。
func (s *StatusPuller) applyResult(ctx context.Context, task *model.PushTask, pa *model.ProviderAccount, pushLog *model.PushLog, r *sender.StatusQueryResult, now time.Time, delivered bool) {
	newStatus := string(model.PushTaskStatusSuccess)
	event := model.WebhookEventDelivered
	callbackStatus := "success"
	if !delivered {
		newStatus = string(model.PushTaskStatusFailed)
		event = model.WebhookEventFailed
		callbackStatus = "failed"
	}

	// CAS：仅 sending 且无回执（空或 timeout）才更新，避免覆盖并发到达的真实回调。
	// 放宽到含 timeout 状态，使主动查询能覆盖 sms_timeout_scanner 标记的任务并纠正误判。
	res := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("task_id = ? AND status = ? AND (callback_status IS NULL OR callback_status = '' OR callback_status = ?)",
			task.TaskID, string(model.PushTaskStatusSending), "timeout").
		Updates(map[string]any{
			"status":          newStatus,
			"callback_status": callbackStatus,
			"callback_time":   now,
			"error_msg":       r.ErrorMessage,
			"updated_at":      now,
		})
	if res.Error != nil || res.RowsAffected == 0 {
		return // 已被并发回调等处理
	}
	logger.Infof("status puller: task %s queried -> %s (provider_id=%s)", task.TaskID, newStatus, r.ProviderMsgID)

	// 更新 push_log
	_ = s.svc.DB.WithContext(ctx).Model(&model.PushLog{}).Where("id = ?", pushLog.ID).
		Updates(map[string]any{
			"status":     newStatus,
			"error_code": r.ErrorCode,
			"error_msg":  r.ErrorMessage,
		}).Error

	// 触发 webhook
	wh := newWebhookDispatch(s.svc)
	wh.Dispatch(ctx, task.AppID, task.TaskID, event, map[string]any{
		"task_id":       task.TaskID,
		"app_id":        task.AppID,
		"channel_id":    task.ChannelID,
		"receiver":      task.Receiver,
		"message_type":  task.MessageType,
		"status":        newStatus,
		"provider_id":   r.ProviderMsgID,
		"provider_code": pa.ProviderCode,
		"source":        "status_pull",
		"occurred_at":   now.Format(time.RFC3339),
	}, task.RequestID)
}
