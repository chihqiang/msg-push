package pipeline

import (
	"context"
	"fmt"
	"time"

	"chihqiang/msg-push/core/common"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
)

// ExecuteContext 动作执行上下文。
type ExecuteContext struct {
	Task              *model.PushTask
	ProviderAccountID uint
	ProviderCode      string
	ErrorCode         string
	ErrorMessage      string
	RequestData       string
	ResponseData      string
	ProviderID        string
	Event             string
}

// ExecuteResult 动作执行结果。
type ExecuteResult struct {
	Action      string
	ShouldRetry bool
	RetryDelay  time.Duration
	Err         error
}

// ActionExecutor 失败动作执行器：按规则动作执行 retry / switch_provider / fail / alert。
type ActionExecutor struct {
	svc      *svc.ServiceContext
	term     *TerminalService
	producer queueEnqueuer
}

// queueEnqueuer 入队接口（便于测试）。
type queueEnqueuer interface {
	Enqueue(ctx context.Context, task *model.PushTask, scheduledAt *time.Time) error
}

// NewActionExecutor 创建动作执行器。
func NewActionExecutor(s *svc.ServiceContext) *ActionExecutor {
	return &ActionExecutor{
		svc:      s,
		term:     NewTerminalService(s),
		producer: &asynqEnqueuer{s: s},
	}
}

// asynqEnqueuer 基于 EnqueueSendMessage 的入队实现。
type asynqEnqueuer struct{ s *svc.ServiceContext }

// Enqueue 重新入队（scheduledAt 非空走延迟）。
func (e *asynqEnqueuer) Enqueue(ctx context.Context, task *model.PushTask, scheduledAt *time.Time) error {
	params := map[string]string{}
	if task.Params != "" {
		_ = common.UnmarshalJSONString(task.Params, &params)
	}
	_, err := logic.EnqueueSendMessage(ctx, e.s.Producer, logic.SendMessagePayload{
		TaskID:     task.TaskID,
		RequestID:  task.RequestID,
		AppID:      task.AppID,
		ChannelID:  task.ChannelID,
		TemplateID: task.TemplateID,
		Receiver:   task.Receiver,
		Params:     params,
		Signature:  task.Signature,
	}, scheduledAt)
	return err
}

// Execute 执行规则动作。
func (e *ActionExecutor) Execute(ctx context.Context, result *EvaluateResult, execCtx *ExecuteContext) *ExecuteResult {
	if result == nil || execCtx == nil || execCtx.Task == nil {
		return &ExecuteResult{Action: model.RuleActionFail, Err: fmt.Errorf("invalid execute context")}
	}
	action := result.Action
	if action == "" {
		action = model.RuleActionFail
	}

	switch action {
	case model.RuleActionRetry:
		return e.executeRetry(ctx, result, execCtx)
	case model.RuleActionSwitchProvider:
		return e.executeSwitchProvider(ctx, result, execCtx)
	case model.RuleActionAlert:
		return e.executeAlert(ctx, result, execCtx)
	default:
		return e.executeFail(ctx, execCtx)
	}
}

// executeRetry 重试：指数退避延迟入队。
func (e *ActionExecutor) executeRetry(ctx context.Context, result *EvaluateResult, execCtx *ExecuteContext) *ExecuteResult {
	task := execCtx.Task
	cfg := &model.RetryActionConfig{MaxRetry: task.MaxRetry, DelaySeconds: 2, BackoffRate: 2, MaxDelay: 60}
	if result.MatchedRule != nil {
		if rc, err := result.MatchedRule.GetRetryConfig(); err == nil {
			cfg = rc
		}
	}
	if task.RetryCount >= cfg.MaxRetry {
		return e.executeFail(ctx, execCtx)
	}

	delay := backoffDelay(task.RetryCount, cfg)
	task.RetryCount++
	task.Status = model.PushTaskStatusPending
	task.ErrorMsg = ""

	updates := map[string]any{
		"retry_count": task.RetryCount,
		"status":      string(model.PushTaskStatusPending),
		"error_msg":   "",
	}
	if err := e.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("task_id = ?", task.TaskID).Updates(updates).Error; err != nil {
		return &ExecuteResult{Action: model.RuleActionRetry, Err: err}
	}

	e.writeLog(ctx, task, execCtx, "retry", fmt.Sprintf("will retry in %ds", int(delay.Seconds())))

	at := time.Now().Add(delay)
	if err := e.producer.Enqueue(ctx, task, &at); err != nil {
		logger.Errorf("retry enqueue task %s failed: %v", task.TaskID, err)
		// 任务已被置为 pending 但未入队，若放任会永久卡死（asynq 不再投递）。
		// 置 failed 终态避免滞留，并保留原始发送错误作为失败原因。
		return e.executeFail(ctx, execCtx)
	}
	return &ExecuteResult{Action: model.RuleActionRetry, ShouldRetry: true, RetryDelay: delay}
}

// executeSwitchProvider 切换供应商：排除当前供应商后立即重新入队。
func (e *ActionExecutor) executeSwitchProvider(ctx context.Context, result *EvaluateResult, execCtx *ExecuteContext) *ExecuteResult {
	task := execCtx.Task
	cfg := &model.SwitchProviderActionConfig{ExcludeCurrent: true, MaxRetry: 1}
	if result.MatchedRule != nil {
		if sc, err := result.MatchedRule.GetSwitchProviderConfig(); err == nil {
			cfg = sc
		}
	}
	if task.RetryCount >= cfg.MaxRetry {
		return e.executeFail(ctx, execCtx)
	}

	if cfg.ExcludeCurrent && execCtx.ProviderAccountID > 0 {
		task.AddExcludeProviderID(execCtx.ProviderAccountID)
	}
	task.RetryCount++
	task.Status = model.PushTaskStatusPending
	task.ErrorMsg = ""

	updates := map[string]any{
		"retry_count":          task.RetryCount,
		"status":               string(model.PushTaskStatusPending),
		"error_msg":            "",
		"exclude_provider_ids": task.ExcludeProviderIDs,
	}
	if err := e.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("task_id = ?", task.TaskID).Updates(updates).Error; err != nil {
		return &ExecuteResult{Action: model.RuleActionSwitchProvider, Err: err}
	}

	e.writeLog(ctx, task, execCtx, "switch_provider", fmt.Sprintf("excluded provider %d", execCtx.ProviderAccountID))

	// 立即重新入队（不延迟，避免窗口内丢失）
	if err := e.producer.Enqueue(ctx, task, nil); err != nil {
		logger.Errorf("switch provider enqueue task %s failed: %v", task.TaskID, err)
		// 任务已被置为 pending 但未入队，置 failed 终态避免永久滞留
		return e.executeFail(ctx, execCtx)
	}
	return &ExecuteResult{Action: model.RuleActionSwitchProvider, ShouldRetry: true}
}

// executeAlert 告警：发送告警 webhook 后进入失败终态。
func (e *ActionExecutor) executeAlert(ctx context.Context, result *EvaluateResult, execCtx *ExecuteContext) *ExecuteResult {
	alertCfg := &model.AlertActionConfig{AlertLevel: "warning"}
	if result.MatchedRule != nil {
		if ac, err := result.MatchedRule.GetAlertConfig(); err == nil {
			alertCfg = ac
		}
	}
	// 发送告警（同步 HTTP POST，失败不影响终态）
	e.sendAlert(ctx, alertCfg, execCtx)
	return e.executeFail(ctx, execCtx)
}

// executeFail 失败终态。
func (e *ActionExecutor) executeFail(ctx context.Context, execCtx *ExecuteContext) *ExecuteResult {
	task := execCtx.Task
	_, err := e.term.Transition(ctx, TerminalTransition{
		TaskID:       task.TaskID,
		Status:       string(model.PushTaskStatusFailed),
		Event:        "failed",
		ProviderID:   execCtx.ProviderID,
		ErrorCode:    execCtx.ErrorCode,
		ErrorMessage: execCtx.ErrorMessage,
	})
	if err != nil {
		logger.Errorf("transition failed task %s: %v", task.TaskID, err)
		return &ExecuteResult{Action: model.RuleActionFail, Err: err}
	}
	if execCtx.ProviderAccountID > 0 {
		e.writeLog(ctx, task, execCtx, "failed", execCtx.ErrorMessage)
	}
	return &ExecuteResult{Action: model.RuleActionFail}
}

// writeLog 写推送日志。
func (e *ActionExecutor) writeLog(ctx context.Context, task *model.PushTask, execCtx *ExecuteContext, status, errMsg string) {
	log := &model.PushLog{
		RequestID:         task.RequestID,
		TaskID:            task.ID,
		TaskNo:            task.TaskID,
		AppID:             task.AppID,
		ProviderAccountID: execCtx.ProviderAccountID,
		ProviderMsgID:     execCtx.ProviderID,
		Receiver:          task.Receiver,
		IsTest:            task.IsTest,
		Status:            status,
		ProviderResp:      execCtx.ResponseData,
		ErrorCode:         execCtx.ErrorCode,
		ErrorMsg:          errMsg,
	}
	if err := e.svc.DB.WithContext(ctx).Create(log).Error; err != nil {
		logger.Warnf("write push log failed: %v", err)
	}
}

// sendAlert 发送告警 webhook（同步，尽力而为）。
func (e *ActionExecutor) sendAlert(ctx context.Context, cfg *model.AlertActionConfig, execCtx *ExecuteContext) {
	webhookURL := cfg.WebhookURL
	if webhookURL == "" {
		return
	}
	payload := map[string]any{
		"alert_type":    "message_push_failure",
		"task_id":       execCtx.Task.TaskID,
		"app_id":        execCtx.Task.AppID,
		"receiver":      execCtx.Task.Receiver,
		"message_type":  execCtx.Task.MessageType,
		"provider":      execCtx.ProviderCode,
		"error_code":    execCtx.ErrorCode,
		"error_message": execCtx.ErrorMessage,
		"timestamp":     time.Now().Format(time.RFC3339),
	}
	body := common.MarshalJSON(payload)
	if len(body) == 0 {
		return
	}
	_, _, _ = common.PostJSON(ctx, webhookURL, body, 10*time.Second)
}

// backoffDelay 指数退避延迟。
func backoffDelay(retryCount int, cfg *model.RetryActionConfig) time.Duration {
	delay := cfg.DelaySeconds
	for i := 0; i < retryCount && delay < cfg.MaxDelay; i++ {
		delay *= cfg.BackoffRate
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}
	return time.Duration(delay) * time.Second
}
