package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"chihqiang/msg-push/core/common"
	"chihqiang/msg-push/core/sender"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
	"gorm.io/gorm"
)

// MessageHandler 消息投递处理器：消费 msg:send 任务，完成选通道、渲染、发送与状态流转。
type MessageHandler struct {
	svc      *svc.ServiceContext
	selector Selector
	renderer Renderer
	engine   Engine
	executor *ActionExecutor
	term     *TerminalService
}

// NewMessageHandler 创建消息处理器。
func NewMessageHandler(s *svc.ServiceContext) *MessageHandler {
	return &MessageHandler{
		svc:      s,
		selector: NewSelector(s),
		renderer: NewRenderer(),
		engine:   NewEngine(s),
		executor: NewActionExecutor(s),
		term:     NewTerminalService(s),
	}
}

// HandlePayload 处理入队消息（payload 为 SendMessagePayload JSON）。
func (h *MessageHandler) HandlePayload(ctx context.Context, payload []byte) error {
	var p logic.SendMessagePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse send payload: %w", err)
	}
	if p.TaskID == "" {
		return errors.New("empty task_id in payload")
	}
	// 恢复全链路追踪ID到 context，使本消费链路所有日志自动携带 request_id
	if p.RequestID != "" {
		ctx = httpx.ContextWithRequestID(ctx, p.RequestID)
	}
	return h.Handle(ctx, p.TaskID)
}

// Handle 处理任务投递。
func (h *MessageHandler) Handle(ctx context.Context, taskID string) error {
	// 1. 查任务
	var task model.PushTask
	if err := h.svc.DB.WithContext(ctx).Where("task_id = ?", taskID).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warnf("task %s not found, skip", taskID)
			return nil
		}
		return err
	}

	// 2. CAS 抢占：pending→sending（at-least-once 下重复消息直接跳过）
	// 必须检查 Error：DB 瞬时故障时 RowsAffected=0 会被误判为"重复任务"而直接跳过，
	// 导致任务永久滞留 pending。出错时返回 err 触发 asynq 重试。
	casRes := h.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("task_id = ? AND status = ?", taskID, string(model.PushTaskStatusPending)).
		Update("status", string(model.PushTaskStatusSending))
	if casRes.Error != nil {
		return fmt.Errorf("cas claim task %s: %w", taskID, casRes.Error)
	}
	if casRes.RowsAffected == 0 {
		logger.Infof("skip duplicate task %s (not pending)", taskID)
		return nil
	}
	task.Status = model.PushTaskStatusSending

	// 3. 选通道
	node, err := h.selectChannel(ctx, &task)
	if err != nil {
		if task.IsTest {
			// 测试任务不依赖供应商配置：选通道失败也模拟成功，验证完整链路
			return h.handleTestSuccess(ctx, &task, 0)
		}
		// 无可用供应商多为瞬态（熔断/类型不匹配/暂不可用）：有限重试避免消息直接丢失；
		// 超过最大重试次数才置 failed。同时入队失败会兜底置 failed，避免永久滞留。
		h.handleNoProvider(ctx, &task, err.Error())
		return nil
	}
	providerAccount := node.ProviderAccount
	if providerAccount == nil {
		if task.IsTest {
			// 测试任务不依赖供应商账号：直接模拟成功
			return h.handleTestSuccess(ctx, &task, 0)
		}
		h.handleEarlyFailure(ctx, &task, 0, "provider account not found")
		return nil
	}

	// 4. 持久化选中供应商
	_ = h.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("task_id = ?", taskID).Update("provider_account_id", providerAccount.ID).Error
	task.ProviderAccountID = providerAccount.ID

	// 5. 解析签名 + 获取发送器
	providerMeta, ok := sender.GetByCode(providerAccount.ProviderCode)
	if !ok {
		h.handleEarlyFailure(ctx, &task, providerAccount.ID, fmt.Sprintf("provider %s not registered", providerAccount.ProviderCode))
		return nil
	}
	providerSignature, err := h.resolveSignature(ctx, &task, providerAccount.ID, providerMeta.RequiresSignature)
	if err != nil {
		h.handleEarlyFailure(ctx, &task, providerAccount.ID, err.Error())
		return nil
	}
	messageSender, err := sender.DefaultResolver.GetSender(providerAccount.ProviderCode)
	if err != nil {
		h.handleEarlyFailure(ctx, &task, providerAccount.ID, err.Error())
		return nil
	}

	// 6. 参数映射 + 模板渲染
	mappedParams, renderedContent := h.renderMessage(ctx, &task, node)

	// 7. 发送（测试任务不走真实发送，直接模拟成功走完整链路）
	if task.IsTest {
		return h.handleTestSuccess(ctx, &task, providerAccount.ID)
	}

	// 7. 发送
	sendReq := &sender.SendRequest{
		Task:                   &task,
		ProviderAccount:        providerAccount,
		ChannelTemplateBinding: node.ChannelTemplateBinding,
		Signature:              providerSignature,
		MappedParams:           mappedParams,
		RenderedContent:        renderedContent,
	}
	if task.MessageType == string(model.ChannelTypeSMS) {
		region, national, e164 := parseCNMobile(task.Receiver)
		sendReq.PhoneRegion = region
		sendReq.PhoneNationalNumber = national
		sendReq.PhoneE164 = e164
	}

	resp, err := messageSender.Send(ctx, sendReq)
	if err != nil {
		logger.Errorf("sender error task_id=%s: %v", taskID, err)
		if resp != nil {
			h.handleSendError(ctx, &task, providerAccount.ID, resp)
		} else {
			h.handleEarlyFailure(ctx, &task, providerAccount.ID, err.Error())
		}
		return nil
	}
	// 发送器异常返回 nil 响应时按早期失败处理，避免 panic
	if resp == nil {
		h.handleEarlyFailure(ctx, &task, providerAccount.ID, "sender returned nil response")
		return nil
	}

	// 8. 处理结果
	if resp.Success {
		return h.handleSuccess(ctx, &task, providerAccount.ID, resp)
	}
	h.handleSendError(ctx, &task, providerAccount.ID, resp)
	return nil
}

// selectChannel 选择发送通道。
func (h *MessageHandler) selectChannel(ctx context.Context, task *model.PushTask) (*ChannelNode, error) {
	return h.selector.SelectWithExcludes(ctx, task.ChannelID, task.MessageType,
		fmt.Sprintf("%d", task.AppID), task.Receiver, task.GetExcludeProviderIDs())
}

// resolveSignature 解析签名映射。
func (h *MessageHandler) resolveSignature(ctx context.Context, task *model.PushTask, providerAccountID uint, requires bool) (*model.ProviderSignature, error) {
	alias := strings.TrimSpace(task.Signature)
	if alias == "" {
		if requires {
			return nil, errors.New("required signature is missing")
		}
		return nil, nil
	}
	var mapping model.ChannelSignatureMapping
	err := h.svc.DB.WithContext(ctx).
		Preload("ProviderSignature").
		Where("channel_id = ? AND signature_name = ? AND provider_id = ? AND status = 1",
			task.ChannelID, alias, providerAccountID).
		First(&mapping).Error
	if err != nil {
		if requires {
			return nil, fmt.Errorf("required signature alias cannot be resolved: %w", err)
		}
		return nil, nil
	}
	if mapping.ProviderSignature == nil {
		return nil, nil
	}
	return mapping.ProviderSignature, nil
}

// renderMessage 参数映射 + 模板渲染。
func (h *MessageHandler) renderMessage(ctx context.Context, task *model.PushTask, node *ChannelNode) (map[string]string, string) {
	// 解析任务参数
	var templateParams map[string]string
	if task.Params != "" {
		_ = json.Unmarshal([]byte(task.Params), &templateParams)
	}

	var mappedParams map[string]string
	if node.ChannelTemplateBinding != nil {
		paramMapping, err := node.ChannelTemplateBinding.GetParamMapping()
		if err == nil {
			if len(paramMapping) > 0 {
				mappedParams = h.renderer.MapParams(templateParams, paramMapping)
			} else if node.ChannelTemplateBinding.ProviderTemplate != nil {
				// 空映射：仅同名变量自动映射
				if vars, err := node.ChannelTemplateBinding.ProviderTemplate.GetVariables(); err == nil {
					mappedParams = SameNameTemplateParams(templateParams, vars)
				}
			}
		}
	}

	// 渲染供应商模板
	rendered := ""
	if node.ChannelTemplateBinding != nil && node.ChannelTemplateBinding.ProviderTemplate != nil {
		rendered = h.renderer.RenderSimple(node.ChannelTemplateBinding.ProviderTemplate.TemplateContent, mappedParams)
	}
	return mappedParams, rendered
}

// handleTestSuccess 测试任务：走完整链路（选通道/渲染/落库/状态流转/批次计数）但模拟发送成功，
// 不调用真实 sender，避免测试请求产生真实外部投递。
func (h *MessageHandler) handleTestSuccess(ctx context.Context, task *model.PushTask, providerAccountID uint) error {
	logger.Infof("test task %s: simulating send success (no real provider call)", task.TaskID)
	testResp := &sender.SendResponse{
		Success:      true,
		Status:       string(model.PushTaskStatusSuccess),
		TaskID:       task.TaskID,
		ProviderID:   "TEST",
		RequestData:  "{}",
		ResponseData: `{"test":true}`,
	}
	return h.handleSuccess(ctx, task, providerAccountID, testResp)
}

// handleSuccess 处理发送成功。
func (h *MessageHandler) handleSuccess(ctx context.Context, task *model.PushTask, providerAccountID uint, resp *sender.SendResponse) error {
	now := time.Now()
	if resp.Status == string(model.PushTaskStatusSuccess) {
		// 直接成功（邮件/企微/钉钉等）
		if _, err := h.term.Transition(ctx, TerminalTransition{
			TaskID:       task.TaskID,
			Status:       string(model.PushTaskStatusSuccess),
			Event:        "success",
			ProviderID:   resp.ProviderID,
			ErrorMessage: "",
		}); err != nil {
			return err
		}
	} else {
		// 等待回执（短信）：置 sending 保持，记录发送时间（供状态主动查询），写日志
		if err := h.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
			Where("task_id = ?", task.TaskID).Updates(map[string]any{
			"status":     string(model.PushTaskStatusSending),
			"sent_at":    now,
			"updated_at": now,
		}).Error; err != nil {
			return err
		}
	}

	// 写日志：短信提交成功（等待回执，resp.Status=sending）用 sending 状态，
	// 避免未收到回执即记 success 导致成功数虚高；终态由回执/主动拉取/超时扫描更新。
	logStatus := "success"
	if resp.Status == string(model.PushTaskStatusSending) {
		logStatus = string(model.PushTaskStatusSending)
	}
	h.writeLog(ctx, task, providerAccountID, resp, logStatus, "")
	h.selector.ReportSuccess(providerAccountID)
	logger.Infof("message sent successfully task_id=%s provider_id=%s status=%s", task.TaskID, resp.ProviderID, resp.Status)
	return nil
}

// handleSendError 处理发送失败（走规则引擎）。
func (h *MessageHandler) handleSendError(ctx context.Context, task *model.PushTask, providerAccountID uint, resp *sender.SendResponse) {
	h.selector.ReportFailure(providerAccountID)

	providerCode := ""
	if pa, err := h.getProviderAccount(ctx, providerAccountID); err == nil {
		providerCode = pa.ProviderCode
	}

	evalReq := &EvaluateRequest{
		Scene:        model.RuleSceneSendFailure,
		ProviderCode: providerCode,
		MessageType:  task.MessageType,
		ErrorCode:    resp.ErrorCode,
		ErrorMessage: resp.ErrorMessage,
		Task:         task,
	}
	evalResult := h.engine.Evaluate(ctx, evalReq)

	execCtx := &ExecuteContext{
		Task:              task,
		ProviderAccountID: providerAccountID,
		ProviderCode:      providerCode,
		ErrorCode:         resp.ErrorCode,
		ErrorMessage:      resp.ErrorMessage,
		RequestData:       resp.RequestData,
		ResponseData:      resp.ResponseData,
		ProviderID:        resp.ProviderID,
		Event:             "failed",
	}
	result := h.executor.Execute(ctx, evalResult, execCtx)
	if result != nil && result.Err != nil {
		logger.Errorf("rule action execute failed task_id=%s: %v", task.TaskID, result.Err)
	}
}

// handleNoProvider 选通道失败处理（无可用供应商）。
// 无可用供应商多为瞬态（供应商熔断、类型不匹配、暂不可用），在重试上限内
// 指数退避延迟重新入队，避免直接置 failed 导致消息永久丢失；超过上限才置 failed。
func (h *MessageHandler) handleNoProvider(ctx context.Context, task *model.PushTask, errorMsg string) {
	maxRetry := task.MaxRetry
	if maxRetry <= 0 {
		maxRetry = 3
	}
	if task.RetryCount < maxRetry {
		task.RetryCount++
		delay := backoffDelay(task.RetryCount, &model.RetryActionConfig{
			MaxRetry: maxRetry, DelaySeconds: 5, BackoffRate: 2, MaxDelay: 300,
		})

		// 置回 pending 以便下次 CAS 抢占，并记录重试次数与原因
		_ = h.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
			Where("task_id = ?", task.TaskID).
			Updates(map[string]any{
				"retry_count": task.RetryCount,
				"status":      string(model.PushTaskStatusPending),
				"error_msg":   errorMsg,
			}).Error

		var params map[string]string
		if task.Params != "" {
			_ = json.Unmarshal([]byte(task.Params), &params)
		}
		at := time.Now().Add(delay)
		if _, err := logic.EnqueueSendMessage(ctx, h.svc.Producer, logic.SendMessagePayload{
			TaskID:     task.TaskID,
			RequestID:  task.RequestID,
			AppID:      task.AppID,
			ChannelID:  task.ChannelID,
			TemplateID: task.TemplateID,
			Receiver:   task.Receiver,
			Params:     params,
			Signature:  task.Signature,
		}, &at); err != nil {
			// 入队失败：任务已置 pending 但未入队，置 failed 避免永久滞留
			logger.Errorf("no provider retry enqueue task %s failed: %v", task.TaskID, err)
			h.handleEarlyFailure(ctx, task, 0, "requeue failed: "+err.Error())
			return
		}
		logger.Warnf("channel unavailable task=%s retry %d/%d in %s: %s",
			task.TaskID, task.RetryCount, maxRetry, delay.String(), errorMsg)
		return
	}
	// 超过重试上限 → 失败终态
	h.handleEarlyFailure(ctx, task, 0, errorMsg)
}

// handleEarlyFailure 处理早期失败（发送前错误，不走规则引擎）。
func (h *MessageHandler) handleEarlyFailure(ctx context.Context, task *model.PushTask, providerAccountID uint, errorMsg string) {
	_, err := h.term.Transition(ctx, TerminalTransition{
		TaskID:       task.TaskID,
		Status:       string(model.PushTaskStatusFailed),
		Event:        "failed",
		ErrorMessage: errorMsg,
	})
	if err != nil {
		logger.Errorf("transition early failure task_id=%s: %v", task.TaskID, err)
	}
	if providerAccountID > 0 {
		h.svc.DB.WithContext(ctx).Create(&model.PushLog{
			RequestID:         task.RequestID,
			TaskID:            task.ID,
			TaskNo:            task.TaskID,
			AppID:             task.AppID,
			ProviderAccountID: providerAccountID,
			Receiver:          task.Receiver,
			IsTest:            task.IsTest,
			Status:            "failed",
			ProviderResp:      "{}",
			ErrorMsg:          errorMsg,
		})
	}
	logger.Errorf("message early failed task_id=%s error=%s", task.TaskID, errorMsg)
}

// getProviderAccount 查询服务商账号。
func (h *MessageHandler) getProviderAccount(ctx context.Context, id uint) (*model.ProviderAccount, error) {
	var pa model.ProviderAccount
	if err := h.svc.DB.WithContext(ctx).Where("id = ?", id).First(&pa).Error; err != nil {
		return nil, err
	}
	return &pa, nil
}

// writeLog 写发送成功日志。
func (h *MessageHandler) writeLog(ctx context.Context, task *model.PushTask, providerAccountID uint, resp *sender.SendResponse, status, errMsg string) {
	h.svc.DB.WithContext(ctx).Create(&model.PushLog{
		RequestID:         task.RequestID,
		TaskID:            task.ID,
		TaskNo:            task.TaskID,
		AppID:             task.AppID,
		ProviderAccountID: providerAccountID,
		ProviderMsgID:     resp.ProviderID,
		Receiver:          task.Receiver,
		IsTest:            task.IsTest,
		Status:            status,
		ProviderResp:      resp.ResponseData,
		ErrorCode:         resp.ErrorCode,
		ErrorMsg:          errMsg,
	})
}

// parseCNMobile 解析手机号：返回 (region, national, e164)。
//   - 中国大陆 11 位号码：region=CN, national=原号, e164=+86 前缀
//   - 国际号码（E.164 形式 + 或 00 开头，或已带国家码）：region=空, national=空, e164=规范化号码
//   - 其他：全部返回空
func parseCNMobile(receiver string) (region, national, e164 string) {
	if common.CNMobileRe.MatchString(receiver) {
		return "CN", receiver, "+86" + receiver
	}
	// 国际号码规范化：支持 +86138...、0086138...、86138... 等常见形式
	if s := normalizeE164(receiver); s != "" {
		return "", "", s
	}
	return "", "", ""
}

// normalizeE164 规范化国际号码为 +<国家码><号码>；无法识别时返回空。
func normalizeE164(receiver string) string {
	s := strings.TrimSpace(receiver)
	if s == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(s, "+"):
		s = s[1:]
	case strings.HasPrefix(s, "00"):
		s = s[2:]
	}
	// 去除空格/短横线/括号等分隔符
	s = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(s)
	if len(s) < 8 || len(s) > 15 || !common.IsAllDigits(s) {
		return ""
	}
	return "+" + s
}
