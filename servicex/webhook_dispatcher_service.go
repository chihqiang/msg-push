package servicex

import (
	"chihqiang/msg-push/webhookx"
)

// WebhookDispatcherService 适配 webhookx.Dispatcher 为 service.Service。
type WebhookDispatcherService struct {
	d *webhookx.Dispatcher
}

// NewWebhookDispatcherService 创建 Webhook 投递器适配器。
func NewWebhookDispatcherService(d *webhookx.Dispatcher) WebhookDispatcherService {
	return WebhookDispatcherService{d: d}
}

func (s WebhookDispatcherService) Start() { s.d.StartService() }

func (s WebhookDispatcherService) Stop() { s.d.Stop() }
