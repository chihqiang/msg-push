package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/sender"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
)

// CallbackRequest 服务商回调请求。
type CallbackRequest struct {
	ProviderAccountID uint
	RawBody           []byte
	Headers           map[string]string
	QueryParams       map[string]string
	FormData          map[string]string
}

// CallbackHandleResult 回调处理结果（服务商期望的响应）。
type CallbackHandleResult struct {
	StatusCode int
	Body       string
}

// parsedCallback 通用解析后的回调数据。
type parsedCallback struct {
	Type         string // report / upstream
	ProviderID   string // 服务商消息ID
	Status       string // delivered / failed / rejected
	ErrorCode    string
	ErrorMessage string
	Mobile       string
	Content      string
}

// CallbackLogic 服务商回调接收与回执处理逻辑。
type CallbackLogic struct {
	svc     *svc.ServiceContext
	webhook *WebhookLogic
}

// NewCallbackLogic 创建回调处理逻辑。
func NewCallbackLogic(s *svc.ServiceContext) *CallbackLogic {
	return &CallbackLogic{svc: s, webhook: NewWebhookLogic(s)}
}

// Handle 处理服务商回调：解析 → 落库 callback_log → 尽力关联更新 push_log/push_task → 触发 webhook。
// 解析策略：优先使用服务商定制解析器（阿里/腾讯/网易等），否则兜底通用解析。
func (l *CallbackLogic) Handle(ctx context.Context, req *CallbackRequest) CallbackHandleResult {
	// 默认响应
	okResp := CallbackHandleResult{StatusCode: 200, Body: `{"code":0,"message":"ok"}`}

	// 1. 查找服务商账号（获取 provider_code 与归属账号）
	var account model.ProviderAccount
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", req.ProviderAccountID).First(&account).Error; err != nil {
		logger.Warnf("callback: provider account not found, id=%d", req.ProviderAccountID)
		return CallbackHandleResult{StatusCode: 200, Body: `{"code":404,"message":"account not found"}`}
	}

	// 2. 解析回调数据：优先服务商定制解析，否则通用解析
	parsedList, resp := l.parseForProvider(ctx, &account, req)
	if resp != nil {
		return *resp
	}

	// 3. 逐个结果落库 + 关联 + 通知
	for _, pc := range parsedList {
		l.processResult(ctx, &account, req, pc)
	}

	return okResp
}

// parseForProvider 按服务商解析回调，返回结果列表；resp 非空表示需要直接返回该响应。
func (l *CallbackLogic) parseForProvider(ctx context.Context, account *model.ProviderAccount, req *CallbackRequest) ([]*parsedCallback, *CallbackHandleResult) {
	// 尝试服务商定制解析器
	if parser := sender.GetCallbackParser(account.ProviderCode); parser != nil {
		senderReq := &sender.CallbackRequest{
			ProviderCode:      account.ProviderCode,
			ProviderAccountID: account.ID,
			RawBody:           req.RawBody,
			Headers:           req.Headers,
			QueryParams:       req.QueryParams,
			FormData:          req.FormData,
		}
		senderResp, results, err := parser.Parse(ctx, senderReq)
		if err != nil {
			logger.Warnf("callback: provider %s parse failed: %v", account.ProviderCode, err)
		}
		if len(results) == 0 && err == nil {
			// 合法但无结果（如网易未知 eventType）：返回服务商期望的响应
			return nil, &CallbackHandleResult{StatusCode: senderResp.StatusCode, Body: senderResp.Body}
		}
		parsed := make([]*parsedCallback, 0, len(results))
		for _, r := range results {
			parsed = append(parsed, &parsedCallback{
				Type:         r.Type,
				ProviderID:   r.ProviderID,
				Status:       r.Status,
				ErrorCode:    r.ErrorCode,
				ErrorMessage: r.ErrorMessage,
				Mobile:       r.Mobile,
				Content:      r.Content,
			})
		}
		if len(parsed) > 0 {
			return parsed, nil
		}
	}

	// 兜底：通用解析
	pc := parseCallback(req.RawBody, req.FormData, req.QueryParams)
	return []*parsedCallback{pc}, nil
}

// processResult 处理单个解析结果：落库 + 关联 + 通知。
func (l *CallbackLogic) processResult(ctx context.Context, account *model.ProviderAccount, req *CallbackRequest, pc *parsedCallback) {
	// 落库回调日志
	log := &model.CallbackLog{
		Type:              model.CallbackType(pc.Type),
		ProviderCode:      account.ProviderCode,
		ProviderAccountID: account.ID,
		ProviderID:        pc.ProviderID,
		Mobile:            pc.Mobile,
		Content:           pc.Content,
		CallbackStatus:    pc.Status,
		ErrorCode:         pc.ErrorCode,
		ErrorMessage:      pc.ErrorMessage,
		RawData:           string(req.RawBody),
	}
	if log.Type == "" {
		log.Type = model.CallbackTypeReport
	}
	if err := l.svc.DB.WithContext(ctx).Create(log).Error; err != nil {
		logger.Errorf("callback: persist callback log failed: %v", err)
	}

	// 尽力关联并更新 push_log / push_task
	taskNo := l.resolveTask(ctx, account, pc, log)

	// 触发 webhook 通知
	l.notifyWebhook(ctx, account, taskNo, log, pc)
}

// resolveTask 尽力关联 push_log（按 provider_msg_id + 手机号），更新状态并回填 task_no。
func (l *CallbackLogic) resolveTask(ctx context.Context, account *model.ProviderAccount, pc *parsedCallback, log *model.CallbackLog) string {
	if pc.ProviderID == "" {
		return ""
	}

	var pushLog model.PushLog
	// 优先精确匹配 (provider_msg_id + receiver)，避免批量场景（同 BizId 多条记录）关联错任务；
	// 仅当回调未携带手机号时才回退为按 provider_msg_id 关联。
	if pc.Mobile != "" {
		err := l.svc.DB.WithContext(ctx).
			Where("provider_msg_id = ? AND receiver = ?", pc.ProviderID, pc.Mobile).
			Order("id DESC").First(&pushLog).Error
		if err != nil {
			logger.Warnf("callback: push_log not found for provider_id=%s receiver=%s", pc.ProviderID, pc.Mobile)
			return ""
		}
	} else {
		if err := l.svc.DB.WithContext(ctx).
			Where("provider_msg_id = ?", pc.ProviderID).
			Order("id DESC").First(&pushLog).Error; err != nil {
			logger.Warnf("callback: push_log not found for provider_id=%s", pc.ProviderID)
			return ""
		}
	}

	// 回填 task_no 与 request_id 到回调日志（全链路追踪关联）
	if (pushLog.TaskNo != "" && log.TaskNo == "") || (pushLog.RequestID != "" && log.RequestID == "") {
		updates := map[string]any{}
		if pushLog.TaskNo != "" && log.TaskNo == "" {
			updates["task_no"] = pushLog.TaskNo
			log.TaskNo = pushLog.TaskNo
			log.AppID = pushLog.AppID
		}
		if pushLog.RequestID != "" && log.RequestID == "" {
			updates["request_id"] = pushLog.RequestID
			log.RequestID = pushLog.RequestID
		}
		if len(updates) > 0 {
			_ = l.svc.DB.WithContext(ctx).Model(&model.CallbackLog{}).
				Where("id = ?", log.ID).Updates(updates).Error
		}
	}

	// 更新 push_log 状态
	if pc.Status == model.CallbackStatusDelivered || pc.Status == model.CallbackStatusFailed || pc.Status == model.CallbackStatusRejected {
		newStatus := string(model.PushTaskStatusSuccess)
		if pc.Status != model.CallbackStatusDelivered {
			newStatus = string(model.PushTaskStatusFailed)
		}
		if pushLog.Status != newStatus {
			_ = l.svc.DB.WithContext(ctx).Model(&model.PushLog{}).
				Where("id = ?", pushLog.ID).
				Updates(map[string]any{
					"status":     newStatus,
					"error_code": pc.ErrorCode,
					"error_msg":  pc.ErrorMessage,
				}).Error
		}
		// 同步更新 push_task 状态（补齐回执状态与时间，与主动拉取/超时扫描口径一致）
		now := time.Now()
		callbackStatus := "success"
		if newStatus == string(model.PushTaskStatusFailed) {
			callbackStatus = "failed"
		}
		_ = l.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
			Where("task_id = ?", pushLog.TaskNo).
			Updates(map[string]any{
				"status":          newStatus,
				"callback_status": callbackStatus,
				"callback_time":   now,
				"error_msg":       pc.ErrorMessage,
				"updated_at":      now,
			}).Error
	}

	return pushLog.TaskNo
}

// notifyWebhook 触发 webhook 通知（outbox 入队，由 Dispatcher 异步投递）。
func (l *CallbackLogic) notifyWebhook(ctx context.Context, account *model.ProviderAccount, taskNo string, log *model.CallbackLog, pc *parsedCallback) {
	event := ""
	status := ""
	switch pc.Status {
	case model.CallbackStatusDelivered:
		event = model.WebhookEventDelivered
		status = string(model.PushTaskStatusSuccess)
	case model.CallbackStatusFailed, model.CallbackStatusRejected:
		event = model.WebhookEventFailed
		status = string(model.PushTaskStatusFailed)
	default:
		return // 未知状态不通知
	}

	payload := map[string]any{
		"event":       event,
		"task_no":     taskNo,
		"app_id":      log.AppID,
		"status":      status,
		"provider_id": pc.ProviderID,
		"mobile":      log.Mobile,
		"error_code":  pc.ErrorCode,
		"error_msg":   pc.ErrorMessage,
		"timestamp":   log.CreatedAt.Unix(),
	}
	l.webhook.Dispatch(ctx, log.AppID, taskNo, event, payload, log.RequestID)
}

// parseCallback 通用解析服务商回调（尽力提取，兼容 JSON 与表单）。
func parseCallback(rawBody []byte, form, query map[string]string) *parsedCallback {
	pc := &parsedCallback{}

	// 优先从 JSON 提取
	if len(rawBody) > 0 {
		var m map[string]any
		if err := json.Unmarshal(rawBody, &m); err == nil {
			pc.ProviderID = firstStr(m, "message_id", "msg_id", "provider_id", "sid", "bizId")
			pc.Status = normalizeStatus(firstStr(m, "status", "report_status", "CallbackStatus"))
			pc.ErrorCode = firstStr(m, "err_code", "error_code", "errCode")
			pc.ErrorMessage = firstStr(m, "err_msg", "error_msg", "errmsg", "message")
			pc.Mobile = firstStr(m, "mobile", "phone", "phone_number")
			pc.Content = firstStr(m, "content", "text")
			// 数组形态的回执（如 aliyun sms_deliver）
			if arr, ok := m["receipts"].([]any); ok && len(arr) > 0 {
				if item, ok := arr[0].(map[string]any); ok {
					pc.ProviderID = firstStr(item, "message_id", "msg_id", "provider_id", "bizId")
					pc.Status = normalizeStatus(firstStr(item, "status", "report_status"))
					pc.ErrorCode = firstStr(item, "err_code", "error_code")
					pc.ErrorMessage = firstStr(item, "err_msg", "error_msg")
					pc.Mobile = firstStr(item, "mobile", "phone")
				}
			}
		}
	}

	// 表单兜底
	for _, f := range []string{"message_id", "msg_id", "provider_id", "sid", "bizId"} {
		if pc.ProviderID == "" {
			pc.ProviderID = firstNonEmpty(f, form, query)
		}
	}
	if pc.Status == "" {
		for _, f := range []string{"status", "report_status"} {
			if pc.Status == "" {
				if v := firstNonEmpty(f, form, query); v != "" {
					pc.Status = normalizeStatus(v)
				}
			}
		}
	}
	if pc.ErrorCode == "" {
		pc.ErrorCode = firstNonEmpty("err_code", form, query)
		if pc.ErrorCode == "" {
			pc.ErrorCode = firstNonEmpty("error_code", form, query)
		}
	}
	if pc.ErrorMessage == "" {
		pc.ErrorMessage = firstNonEmpty("err_msg", form, query)
		if pc.ErrorMessage == "" {
			pc.ErrorMessage = firstNonEmpty("error_msg", form, query)
		}
	}
	if pc.Mobile == "" {
		pc.Mobile = firstNonEmpty("mobile", form, query)
		if pc.Mobile == "" {
			pc.Mobile = firstNonEmpty("phone", form, query)
		}
	}

	return pc
}

// normalizeStatus 规范化回调状态。
func normalizeStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case s == "":
		return ""
	case s == "undelivered":
		// 注意：undelivered 含 "deliver" 子串，必须先于 delivered 判断，否则误判为已送达
		return model.CallbackStatusFailed
	case strings.Contains(s, "deliver") || s == "success" || s == "sent":
		return model.CallbackStatusDelivered
	case s == "failed" || s == "fail":
		return model.CallbackStatusFailed
	case s == "rejected" || s == "refuse":
		return model.CallbackStatusRejected
	default:
		return model.CallbackStatusUnknown
	}
}

// firstStr 从 map 中按 key 顺序取第一个非空字符串值。
func firstStr(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
			if n, ok := v.(float64); ok && n != 0 {
				return fmt.Sprintf("%.0f", n)
			}
		}
	}
	return ""
}

// firstNonEmpty 从多个 map 中按 key 取第一个非空值。
func firstNonEmpty(key string, srcs ...map[string]string) string {
	for _, src := range srcs {
		if v, ok := src[key]; ok && v != "" {
			return v
		}
	}
	return ""
}
