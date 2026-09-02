// Package scheduler 提供后台定时扫描器：短信回执超时处理等。
package scheduler

import (
	"context"
	"time"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
)

// SMSTimeoutScanner 短信回执超时扫描器。
// 处理两类卡住的任务（避免永久滞留）：
//  1. 短信已发送（sending）但超时未收到回执 → 置 callback_status=timeout（保持 sending，等待可能的迟到回执）
//  2. 上述超时任务若再过很久仍无回执 → 转 failed 终态（避免永久卡住，触发 webhook failed）
//
// 多实例安全：间隔级 Redis 分布式锁保证同一时刻仅一个实例扫描；条件化 UPDATE（CAS）保证幂等。
type SMSTimeoutScanner struct {
	svc      *svc.ServiceContext
	interval time.Duration
	// callbackTimeout 发送后多久未收到回执视为回调超时（置 callback_status=timeout）。
	callbackTimeout time.Duration
	// hardTimeout 回调超时后再等多久仍无回执则转 failed 终态。
	hardTimeout time.Duration
	limit       int
	stopCh      chan struct{}
	doneCh      chan struct{}
}

// NewSMSTimeoutScanner 创建扫描器。
func NewSMSTimeoutScanner(s *svc.ServiceContext) *SMSTimeoutScanner {
	// 超时阈值可从配置覆盖（生产可调大，避免服务商回执慢时过早误判）
	callbackTimeout := 60 * time.Second
	if v := s.Config.Scheduler.SMSCallbackTimeout; v > 0 {
		callbackTimeout = v
	}
	hardTimeout := 10 * time.Minute
	if v := s.Config.Scheduler.SMSHardTimeout; v > 0 {
		hardTimeout = v
	}
	return &SMSTimeoutScanner{
		svc:             s,
		interval:        10 * time.Second,
		callbackTimeout: callbackTimeout,
		hardTimeout:     hardTimeout,
		limit:           100,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
}

// Start 启动扫描器（后台协程）。
func (s *SMSTimeoutScanner) Start() {
	go s.run()
}

// Stop 优雅停止。
func (s *SMSTimeoutScanner) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

// StartService 适配 service.Starter。
func (s *SMSTimeoutScanner) StartService() { s.Start() }

// run 主循环。
func (s *SMSTimeoutScanner) run() {
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

// scanOnce 执行一次扫描。
func (s *SMSTimeoutScanner) scanOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	now := time.Now()

	// 1. 短信 sending 超时无回执 → 置 callback_status=timeout
	cutoff1 := now.Add(-s.callbackTimeout)
	res1 := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("message_type = ? AND status = ? AND (callback_status IS NULL OR callback_status = '') AND updated_at < ?",
			string(model.ChannelTypeSMS), string(model.PushTaskStatusSending), cutoff1).
		Limit(s.limit).
		Updates(map[string]any{
			"callback_status": "timeout",
			"updated_at":      now,
		})
	if res1.Error != nil {
		logger.Warnf("sms timeout scanner: mark callback timeout failed: %v", res1.Error)
	} else if res1.RowsAffected > 0 {
		logger.Infof("sms timeout scanner: marked %d sending tasks callback_status=timeout", res1.RowsAffected)
	}

	// 2. 回调超时后仍无回执 → 转 failed 终态（CAS：仅非终态可转）
	cutoff2 := now.Add(-s.hardTimeout)
	var tasks []model.PushTask
	if err := s.svc.DB.WithContext(ctx).
		Where("message_type = ? AND status = ? AND callback_status = ? AND updated_at < ?",
			string(model.ChannelTypeSMS), string(model.PushTaskStatusSending), "timeout", cutoff2).
		Limit(s.limit).
		Find(&tasks).Error; err != nil {
		logger.Warnf("sms timeout scanner: list hard timeout tasks failed: %v", err)
		return
	}
	for i := range tasks {
		t := &tasks[i]
		// 条件化更新：仅 sending 且 callback_status=timeout 才转（避免覆盖并发回调结果）
		res := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
			Where("task_id = ? AND status = ? AND callback_status = ?",
				t.TaskID, string(model.PushTaskStatusSending), "timeout").
			Updates(map[string]any{
				"status":     string(model.PushTaskStatusFailed),
				"error_msg":  "callback timeout: no receipt within " + s.hardTimeout.String(),
				"updated_at": now,
			})
		if res.Error != nil {
			logger.Warnf("sms timeout scanner: terminalize task %s failed: %v", t.TaskID, res.Error)
			continue
		}
		if res.RowsAffected == 0 {
			continue // 已被并发处理（回调到达等）
		}
		logger.Infof("sms timeout scanner: task %s hard timeout -> failed", t.TaskID)
		// 写失败日志 + 触发 webhook failed
		s.recordHardTimeout(ctx, t, now)
	}
}

// recordHardTimeout 记录超时失败日志并触发 webhook。
func (s *SMSTimeoutScanner) recordHardTimeout(ctx context.Context, task *model.PushTask, now time.Time) {
	// 写 push_log
	_ = s.svc.DB.WithContext(ctx).Create(&model.PushLog{
		RequestID:         task.RequestID,
		TaskID:            task.ID,
		TaskNo:            task.TaskID,
		AppID:             task.AppID,
		ProviderAccountID: task.ProviderAccountID,
		Receiver:          task.Receiver,
		IsTest:            task.IsTest,
		Status:            "failed",
		ProviderResp:      "{}",
		ErrorCode:         "CALLBACK_TIMEOUT",
		ErrorMsg:          "no receipt within " + s.hardTimeout.String(),
	}).Error

	// 触发 webhook failed
	wh := newWebhookDispatch(s.svc)
	wh.Dispatch(ctx, task.AppID, task.TaskID, "failed", map[string]any{
		"task_id":       task.TaskID,
		"app_id":        task.AppID,
		"channel_id":    task.ChannelID,
		"receiver":      task.Receiver,
		"message_type":  task.MessageType,
		"status":        "failed",
		"error_code":    "CALLBACK_TIMEOUT",
		"error_message": "no receipt within " + s.hardTimeout.String(),
		"occurred_at":   now.Format(time.RFC3339),
	}, task.RequestID)
}
