package pipeline

import (
	"context"

	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
	"github.com/chihqiang/infra-go/taskq"
	"github.com/hibiken/asynq"
)

// Consumer 消费端：消费 msg:send / msg:batch 队列。
type Consumer struct {
	consumer *taskq.Consumer
	handler  *MessageHandler
	batch    *BatchMessageHandler
}

// NewConsumer 创建消费端（注册 msg:send / msg:batch 处理器）。
func NewConsumer(s *svc.ServiceContext) *Consumer {
	handler := NewMessageHandler(s)
	batch := NewBatchMessageHandler(s)
	consumer := taskq.NewConsumer(taskq.Config{
		RedisAddr:       s.Config.Taskq.RedisAddr,
		RedisPassword:   s.Config.Taskq.RedisPassword,
		RedisDB:         s.Config.Taskq.RedisDB,
		DefaultMaxRetry: s.Config.Taskq.DefaultMaxRetry,
		DefaultTimeout:  s.Config.Taskq.DefaultTimeout,
		DefaultQueue:    s.Config.Taskq.DefaultQueue,
		Concurrency:     s.Config.Taskq.Concurrency,
	}, logger.GetGlobal())

	consumer.HandleFunc(logic.TaskSendMessage, func(ctx context.Context, t *asynq.Task) error {
		return handler.HandlePayload(ctx, t.Payload())
	})
	consumer.HandleFunc(logic.TaskSendBatchMessage, func(ctx context.Context, t *asynq.Task) error {
		return batch.HandlePayload(ctx, t.Payload())
	})

	return &Consumer{consumer: consumer, handler: handler, batch: batch}
}

// Start 启动消费端（非阻塞）。
func (c *Consumer) Start() error {
	return c.consumer.Start()
}

// StartService 适配 service.Starter（无 error 返回）。
func (c *Consumer) StartService() {
	if err := c.Start(); err != nil {
		logger.Errorf("worker consumer start failed: %v", err)
	}
}

// Stop 停止消费端（适配 service.Stopper）。
func (c *Consumer) Stop() {
	c.Shutdown()
}

// Shutdown 优雅关闭。
func (c *Consumer) Shutdown() {
	if c.consumer != nil {
		c.consumer.Shutdown()
	}
}
