package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
	"github.com/chihqiang/infra-go/stringx"
	"gorm.io/gorm"
)

// MessageLogic 消息发送逻辑：校验、落库并入队，真实投递由消费端完成。
type MessageLogic struct {
	svc *svc.ServiceContext
}

// NewMessageLogic 创建消息发送逻辑。
func NewMessageLogic(s *svc.ServiceContext) *MessageLogic {
	return &MessageLogic{svc: s}
}

// Send 创建单条推送任务并入队。batchID 非空表示属于某批量批次。
func (l *MessageLogic) Send(ctx context.Context, app *model.Application, req *dto.SendRequest) (*dto.SendResponse, error) {
	return l.send(ctx, app, req, "")
}

// send 内部实现，支持 batchID。
func (l *MessageLogic) send(ctx context.Context, app *model.Application, req *dto.SendRequest, batchID string) (*dto.SendResponse, error) {
	channel, templateID, err := l.resolveChannelAndTemplate(ctx, req.ChannelCode, req.TemplateCode)
	if err != nil {
		return nil, err
	}

	params, err := marshalParams(req.TemplateParams)
	if err != nil {
		return nil, err
	}

	// 是否测试：直接由应用配置决定（测试应用走完整链路但不真实发送，模拟成功）
	isTest := app.IsTest

	// 全链路追踪ID：来自 HTTP 请求 X-Request-Id，贯穿提交→投递→回执
	requestID := httpx.RequestIDFromContext(ctx)

	task := model.PushTask{
		TaskID:      "task_" + stringx.RandId(),
		RequestID:   requestID,
		AppID:       app.ID,
		BatchID:     batchID,
		ChannelID:   channel.ID,
		TemplateID:  templateID,
		MessageType: string(channel.Type),
		Receiver:    req.Receiver,
		Params:      params,
		Signature:   req.SignatureName,
		IsTest:      isTest,
		Status:      model.PushTaskStatusPending,
		ScheduledAt: req.ScheduledAt,
	}
	if err := l.svc.DB.WithContext(ctx).Create(&task).Error; err != nil {
		return nil, err
	}

	// 入队投递任务
	if _, err := EnqueueSendMessage(ctx, l.svc.Producer, SendMessagePayload{
		TaskID:     task.TaskID,
		RequestID:  task.RequestID,
		AppID:      task.AppID,
		ChannelID:  task.ChannelID,
		TemplateID: task.TemplateID,
		Receiver:   task.Receiver,
		Params:     req.TemplateParams,
		Signature:  task.Signature,
	}, req.ScheduledAt); err != nil {
		// 入队失败：任务标记失败，避免静默丢失
		_ = l.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
			Where("task_id = ?", task.TaskID).
			Updates(map[string]any{
				"status":    model.PushTaskStatusFailed,
				"error_msg": "enqueue failed: " + err.Error(),
			}).Error
		logger.Errorf("enqueue task %s failed: %v", task.TaskID, err)
		return nil, fmt.Errorf("enqueue task failed: %w", err)
	}

	return &dto.SendResponse{
		TaskID:    task.TaskID,
		Status:    string(task.Status),
		IsTest:    isTest,
		CreatedAt: task.CreatedAt,
	}, nil
}

// BatchSend 批量发送：创建批次记录与任务，整批入队一条 msg:batch 由消费端
// 聚合调用服务商批量 API（不支持批量时消费端自动退回逐条发送）。
func (l *MessageLogic) BatchSend(ctx context.Context, app *model.Application, req *dto.BatchSendRequest) (*dto.BatchSendResponse, error) {
	channel, templateID, err := l.resolveChannelAndTemplate(ctx, req.ChannelCode, req.TemplateCode)
	if err != nil {
		return nil, err
	}

	batchID := "batch_" + stringx.RandId()
	total := len(req.Receivers)

	// 是否测试：直接由应用配置决定（测试应用整批模拟成功，不真实发送）
	isTest := app.IsTest

	// 创建批次记录
	batch := &model.PushBatchTask{
		AppID:        app.ID,
		BatchID:      batchID,
		ChannelID:    channel.ID,
		TemplateID:   templateID,
		TotalCount:   total,
		PendingCount: total,
		IsTest:       isTest,
		Status:       model.PushBatchStatusProcessing,
	}
	if err := l.svc.DB.WithContext(ctx).Create(batch).Error; err != nil {
		logger.Errorf("create batch task failed: %v", err)
	}

	// 全链路追踪ID：批量批次内所有子任务共用同一 HTTP 请求的 request_id
	requestID := httpx.RequestIDFromContext(ctx)

	// 为每个接收者创建任务（暂不入队，由 msg:batch 聚合投递）
	// 参数序列化提前到循环外，避免每个接收者重复序列化
	params, perr := marshalParams(req.TemplateParams)
	if perr != nil {
		logger.Warnf("batch send params failed: %v", perr)
	}
	created := 0
	for _, receiver := range req.Receivers {
		task := model.PushTask{
			TaskID:      "task_" + stringx.RandId(),
			RequestID:   requestID,
			AppID:       app.ID,
			BatchID:     batchID,
			ChannelID:   channel.ID,
			TemplateID:  templateID,
			MessageType: string(channel.Type),
			Receiver:    receiver,
			Params:      params,
			Signature:   req.SignatureName,
			IsTest:      isTest,
			Status:      model.PushTaskStatusPending,
			ScheduledAt: req.ScheduledAt,
		}
		if err := l.svc.DB.WithContext(ctx).Create(&task).Error; err != nil {
			logger.Warnf("batch send create task for %s failed: %v", receiver, err)
			continue
		}
		created++
	}

	success := 0
	failed := total - created
	if created > 0 {
		// 整批入队一条 msg:batch（scheduledAt 非空时按计划时间到点整批触发）
		if _, err := EnqueueSendBatchMessage(ctx, l.svc.Producer, SendBatchMessagePayload{BatchID: batchID, RequestID: requestID}, req.ScheduledAt); err != nil {
			logger.Errorf("enqueue batch %s failed: %v", batchID, err)
			// 入队失败：批次下所有任务标记失败，避免永久滞留
			_ = l.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
				Where("batch_id = ? AND status = ?", batchID, string(model.PushTaskStatusPending)).
				Updates(map[string]any{
					"status":    model.PushTaskStatusFailed,
					"error_msg": "enqueue batch failed: " + err.Error(),
				}).Error
			failed = total
		}
		// 注意：创建成功≠发送成功，success 保持 0，由消费端逐条终态流转累加
		// （避免与 TerminalService.updateBatchCount 的 +1 双计数）
	}

	// 更新批次计数与状态
	pending := total - success - failed
	if pending < 0 {
		pending = 0
	}
	batchStatus := model.PushBatchStatusProcessing
	if pending == 0 {
		batchStatus = model.PushBatchStatusCompleted
	}
	_ = l.svc.DB.WithContext(ctx).Model(&model.PushBatchTask{}).
		Where("batch_id = ?", batchID).
		Updates(map[string]any{
			"success_count": success,
			"failed_count":  failed,
			"pending_count": pending,
			"status":        batchStatus,
		}).Error

	return &dto.BatchSendResponse{
		BatchID:      batchID,
		TotalCount:   total,
		SuccessCount: success,
		FailedCount:  failed,
		IsTest:       isTest,
	}, nil
}

// QueryTask 按任务 ID 查询任务详情（限定当前应用所属账号）。
func (l *MessageLogic) QueryTask(ctx context.Context, app *model.Application, taskID string) (*model.PushTask, error) {
	var task model.PushTask
	err := l.svc.DB.WithContext(ctx).
		Where("task_id = ? AND app_id = ?", taskID, app.ID).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// getChannelByCode 校验通道存在且启用（按唯一编码定位）。
func (l *MessageLogic) getChannelByCode(ctx context.Context, code string) (*model.Channel, error) {
	var channel model.Channel
	err := l.svc.DB.WithContext(ctx).
		Where("code = ? AND status = 1", code).
		First(&channel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("channel not found or disabled")
	}
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// getChannelByID 校验通道存在且启用（按主键定位）。
func (l *MessageLogic) getChannelByID(ctx context.Context, id uint) (*model.Channel, error) {
	var channel model.Channel
	err := l.svc.DB.WithContext(ctx).
		Where("id = ? AND status = 1", id).
		First(&channel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("channel not found or disabled")
	}
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// getTemplateByCodeAny 按唯一编码查询启用模板（不限定通道，用于反查模板所属通道）。
func (l *MessageLogic) getTemplateByCodeAny(ctx context.Context, code string) (*model.MessageTemplate, error) {
	var tpl model.MessageTemplate
	err := l.svc.DB.WithContext(ctx).
		Where("code = ? AND status = 1", code).
		First(&tpl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("template not found or disabled")
	}
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// getTemplateByCode 校验模板存在、启用且绑定到指定通道（按唯一编码定位）。
func (l *MessageLogic) getTemplateByCode(ctx context.Context, channelID uint, code string) (*model.MessageTemplate, error) {
	var tpl model.MessageTemplate
	err := l.svc.DB.WithContext(ctx).
		Where("code = ? AND channel_id = ? AND status = 1", code, channelID).
		First(&tpl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("template not found or not bound to channel")
	}
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

// resolveChannelAndTemplate 解析通道与模板，返回通道与模板主键 ID。
//
// 规则：
//   - 只传 template_code：按模板反查其所属通道（channel_code 可省略）；
//   - 传了 channel_code：按编码定位通道，若同时传了 template_code 则校验模板属于该通道；
//   - 两者都传且模板不属于该通道：报错；
//   - 两者都不传：报错。
func (l *MessageLogic) resolveChannelAndTemplate(ctx context.Context, channelCode, templateCode string) (*model.Channel, uint, error) {
	var channel *model.Channel
	if channelCode != "" {
		ch, err := l.getChannelByCode(ctx, channelCode)
		if err != nil {
			return nil, 0, err
		}
		channel = ch
	} else if templateCode != "" {
		// 只传模板：反查模板所属通道
		tpl, err := l.getTemplateByCodeAny(ctx, templateCode)
		if err != nil {
			return nil, 0, err
		}
		ch, err := l.getChannelByID(ctx, tpl.ChannelID)
		if err != nil {
			return nil, 0, err
		}
		channel = ch
	} else {
		return nil, 0, errors.New("channel_code or template_code required")
	}

	templateID := uint(0)
	if templateCode != "" {
		tpl, err := l.getTemplateByCode(ctx, channel.ID, templateCode)
		if err != nil {
			return nil, 0, err
		}
		templateID = tpl.ID
	}
	return channel, templateID, nil
}

// marshalParams 序列化模板参数为 JSON 字符串。
func marshalParams(params map[string]string) (string, error) {
	if len(params) == 0 {
		return "", nil
	}
	b, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("invalid template params: %w", err)
	}
	return string(b), nil
}
