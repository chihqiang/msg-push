package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/queue"
	"chihqiang/msg-push/sender"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
)

// BatchMessageHandler 批量消息投递处理器：消费 msg:batch 任务，
// 聚合批次内所有 pending 任务调用服务商批量 API（一次 HTTP 发送多个号码），
// 逐个结果回填任务。失败/不支持批量时退回逐条入队 msg:send，复用完整重试链路。
type BatchMessageHandler struct {
	svc    *svc.ServiceContext
	single *MessageHandler // 复用选通道/渲染/成功处理
}

// NewBatchMessageHandler 创建批量消息处理器。
func NewBatchMessageHandler(s *svc.ServiceContext) *BatchMessageHandler {
	return &BatchMessageHandler{svc: s, single: NewMessageHandler(s)}
}

// HandlePayload 处理入队消息（payload 为 queue.SendBatchMessagePayload JSON）。
func (h *BatchMessageHandler) HandlePayload(ctx context.Context, payload []byte) error {
	var p queue.SendBatchMessagePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("parse batch payload: %w", err)
	}
	if p.BatchID == "" {
		return errors.New("empty batch_id in payload")
	}
	// 恢复全链路追踪ID到 context，使本批量消费链路日志自动携带 request_id
	if p.RequestID != "" {
		ctx = httpx.ContextWithRequestID(ctx, p.RequestID)
	}
	return h.Handle(ctx, p.BatchID)
}

// Handle 处理批量批次投递。
func (h *BatchMessageHandler) Handle(ctx context.Context, batchID string) error {
	// 0. 防重：Redis 锁保证同一批次同一时刻仅一个实例处理（at-least-once 下防重复批量发送）。
	// 批量任务状态在整批发送后才逐个更新，首次处理若中途崩溃，重复消费可能再次整批发送。
	lockKey := "msgpush:batch_lock:" + batchID
	locked, err := h.svc.Redis.Client().SetNX(ctx, lockKey, 1, 5*time.Minute).Result()
	if err != nil {
		// 锁服务异常时放行（退化为可能重复发送，但不阻塞投递）
		logger.Warnf("batch handler: acquire lock for batch %s failed: %v", batchID, err)
	} else if !locked {
		logger.Infof("batch handler: batch %s already processing, skip", batchID)
		return nil
	} else {
		defer h.svc.Redis.Client().Del(context.Background(), lockKey)
	}

	// 1. 查批次下所有 pending 任务（仅处理待发送的，避免重复消费已完成部分）
	var tasks []model.PushTask
	if err := h.svc.DB.WithContext(ctx).
		Where("batch_id = ? AND status = ?", batchID, string(model.PushTaskStatusPending)).
		Order("id ASC").Find(&tasks).Error; err != nil {
		logger.Warnf("batch handler: list tasks for batch %s failed: %v", batchID, err)
		return err
	}
	if len(tasks) == 0 {
		return nil
	}
	// 转为指针切片（与批量发送请求对齐）
	ptasks := make([]*model.PushTask, len(tasks))
	for i := range tasks {
		ptasks[i] = &tasks[i]
	}

	// 2. 测试批次：不走真实发送，逐任务模拟成功走完整状态流转（不依赖供应商配置）
	if ptasks[0].IsTest {
		logger.Infof("batch handler: batch %s is test, simulating success for %d tasks", batchID, len(ptasks))
		testResp := &sender.SendResponse{
			Success:      true,
			Status:       string(model.PushTaskStatusSuccess),
			ProviderID:   "TEST",
			RequestData:  "{}",
			ResponseData: `{"test":true}`,
		}
		for i := range ptasks {
			if err := h.single.handleSuccess(ctx, ptasks[i], 0, testResp); err != nil {
				logger.Warnf("batch handler: apply test success for task %s failed: %v", ptasks[i].TaskID, err)
			}
		}
		return nil
	}

	// 2. 统一选通道（用第一个任务，批次同通道/模板/签名）
	node, err := h.single.selectChannel(ctx, ptasks[0])
	if err != nil {
		logger.Warnf("batch handler: select channel for batch %s failed: %v, fallback to per-task", batchID, err)
		return h.requeueAll(ctx, ptasks)
	}
	providerAccount := node.ProviderAccount
	if providerAccount == nil {
		logger.Warnf("batch handler: provider account not found for batch %s, fallback to per-task", batchID)
		return h.requeueAll(ctx, ptasks)
	}

	// 3. 解析签名
	providerMeta, ok := sender.GetByCode(providerAccount.ProviderCode)
	if !ok {
		logger.Warnf("batch handler: provider %s not registered, fallback to per-task", providerAccount.ProviderCode)
		return h.requeueAll(ctx, ptasks)
	}
	providerSignature, err := h.single.resolveSignature(ctx, ptasks[0], providerAccount.ID, providerMeta.RequiresSignature)
	if err != nil {
		logger.Warnf("batch handler: resolve signature failed: %v, fallback to per-task", err)
		return h.requeueAll(ctx, ptasks)
	}

	// 4. 获取批量发送器；不支持则退回逐条
	batchSender, err := sender.DefaultResolver.GetBatchSender(providerAccount.ProviderCode)
	if err != nil {
		logger.Infof("batch handler: provider %s does not support batch send, fallback to per-task", providerAccount.ProviderCode)
		return h.requeueAll(ctx, ptasks)
	}

	// 5. 渲染每个任务的参数
	taskParams := make([]map[string]string, len(ptasks))
	rendered := ""
	for i := range ptasks {
		mapped, content := h.single.renderMessage(ctx, ptasks[i], node)
		taskParams[i] = mapped
		if i == 0 {
			rendered = content
		}
	}

	// 6. 批量发送
	batchReq := &sender.BatchSendRequest{
		Tasks:                  ptasks,
		ProviderAccount:        providerAccount,
		ChannelTemplateBinding: node.ChannelTemplateBinding,
		Signature:              providerSignature,
		MappedParams:           taskParams[0],
		TaskParams:             taskParams,
		RenderedContent:        rendered,
	}
	resp, err := batchSender.BatchSend(ctx, batchReq)
	if err != nil {
		logger.Warnf("batch handler: batch send for %s failed: %v, fallback to per-task", batchID, err)
		return h.requeueAll(ctx, ptasks)
	}

	// 7. 逐结果处理：成功 → 更新 sending + 日志；失败 → 退回逐条（复用重试链路）
	if len(resp.Results) != len(ptasks) {
		logger.Warnf("batch handler: result count mismatch (%d vs %d), fallback to per-task", len(resp.Results), len(ptasks))
		return h.requeueAll(ctx, ptasks)
	}
	var fallback []*model.PushTask
	for i, r := range resp.Results {
		task := ptasks[i]
		if r.Success {
			if err := h.single.handleSuccess(ctx, task, providerAccount.ID, r); err != nil {
				logger.Warnf("batch handler: apply success for task %s failed: %v", task.TaskID, err)
				fallback = append(fallback, task)
				continue
			}
			// 持久化选中供应商
			_ = h.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
				Where("task_id = ?", task.TaskID).Update("provider_account_id", providerAccount.ID).Error
			continue
		}
		// 失败任务退回逐条（保持 pending，由单条处理器走规则引擎/重试/切服务商）
		fallback = append(fallback, task)
	}
	if len(fallback) > 0 {
		logger.Infof("batch handler: batch %s fallback %d failed tasks to per-task send", batchID, len(fallback))
		return h.requeueAll(ctx, fallback)
	}
	return nil
}

// requeueAll 将任务退回逐条入队 msg:send（保持 pending，复用完整投递链路）。
// 入队失败的任务标记 failed 避免永久卡住。
func (h *BatchMessageHandler) requeueAll(ctx context.Context, tasks []*model.PushTask) error {
	var firstErr error
	for _, task := range tasks {
		var params map[string]string
		if task.Params != "" {
			_ = json.Unmarshal([]byte(task.Params), &params)
		}
		if _, err := queue.EnqueueSendMessage(ctx, h.svc.Producer, queue.SendMessagePayload{
			TaskID:     task.TaskID,
			RequestID:  task.RequestID,
			AppID:      task.AppID,
			ChannelID:  task.ChannelID,
			TemplateID: task.TemplateID,
			Receiver:   task.Receiver,
			Params:     params,
			Signature:  task.Signature,
		}, nil); err != nil {
			logger.Errorf("batch handler: requeue task %s failed: %v", task.TaskID, err)
			h.single.handleEarlyFailure(ctx, task, 0, "requeue failed: "+err.Error())
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
