package logic

import (
	"context"
	"time"

	"github.com/chihqiang/infra-go/taskq"
	"github.com/hibiken/asynq"
)

// 任务类型常量。
const (
	// TaskSendMessage 消息投递任务。
	TaskSendMessage = "msg:send"
	// TaskSendBatchMessage 批量消息投递任务（批次任务聚合调用服务商批量 API）。
	TaskSendBatchMessage = "msg:batch"
)

// SendMessagePayload 消息投递任务负载，消费端据此完成实际发送。
type SendMessagePayload struct {
	TaskID     string            `json:"task_id"`
	RequestID  string            `json:"request_id"`
	AppID      uint              `json:"app_id"`
	ChannelID  uint              `json:"channel_id"`
	TemplateID uint              `json:"template_id"`
	Receiver   string            `json:"receiver"`
	Params     map[string]string `json:"params"`
	Signature  string            `json:"signature"`
}

// SendBatchMessagePayload 批量消息投递任务负载，消费端据此聚合批次任务调用服务商批量 API。
type SendBatchMessagePayload struct {
	BatchID   string `json:"batch_id"`
	RequestID string `json:"request_id"`
}

// EnqueueSendMessage 将消息投递任务入队。
// scheduledAt 非空时按计划时间投递，否则立即执行。
func EnqueueSendMessage(ctx context.Context, producer *taskq.Producer, p SendMessagePayload, scheduledAt *time.Time) (*asynq.TaskInfo, error) {
	var opts []asynq.Option
	if scheduledAt != nil {
		opts = append(opts, asynq.ProcessAt(*scheduledAt))
	}
	return producer.EnqueuePayload(ctx, TaskSendMessage, p, opts...)
}

// EnqueueSendBatchMessage 将批量消息投递任务入队。
// scheduledAt 非空时按计划时间投递（整批统一到点触发），否则立即执行。
func EnqueueSendBatchMessage(ctx context.Context, producer *taskq.Producer, p SendBatchMessagePayload, scheduledAt *time.Time) (*asynq.TaskInfo, error) {
	var opts []asynq.Option
	if scheduledAt != nil {
		opts = append(opts, asynq.ProcessAt(*scheduledAt))
	}
	return producer.EnqueuePayload(ctx, TaskSendBatchMessage, p, opts...)
}
