package servicex

import (
	"chihqiang/msg-push/worker"
)

// ConsumerService 适配 worker.Consumer 为 service.Service。
type ConsumerService struct {
	c *worker.Consumer
}

// NewConsumerService 创建消费端适配器。
func NewConsumerService(c *worker.Consumer) ConsumerService {
	return ConsumerService{c: c}
}

func (s ConsumerService) Start() { s.c.StartService() }

func (s ConsumerService) Stop() { s.c.Stop() }
