// Package servicex 提供各后台服务的 service.Service 适配器，统一纳入 service.NewServiceGroup 编排。
//
// main 负责装配依赖并构造各后台组件，本包仅做薄适配，使其满足 infra-go service.Service
// 接口（Start/Stop），配合 service.NewServiceGroup 统一启停、优雅关闭。
package servicex

import (
	"chihqiang/msg-push/scheduler"
	"chihqiang/msg-push/webhookx"
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

// SMSTimeoutScannerService 适配 scheduler.SMSTimeoutScanner 为 service.Service。
type SMSTimeoutScannerService struct {
	s *scheduler.SMSTimeoutScanner
}

// NewSMSTimeoutScannerService 创建短信回执超时扫描器适配器。
func NewSMSTimeoutScannerService(sc *scheduler.SMSTimeoutScanner) SMSTimeoutScannerService {
	return SMSTimeoutScannerService{s: sc}
}

func (s SMSTimeoutScannerService) Start() { s.s.StartService() }

func (s SMSTimeoutScannerService) Stop() { s.s.Stop() }

// QuotaSyncerService 适配 scheduler.QuotaSyncer 为 service.Service。
type QuotaSyncerService struct {
	s *scheduler.QuotaSyncer
}

// NewQuotaSyncerService 创建配额统计同步器适配器。
func NewQuotaSyncerService(sync *scheduler.QuotaSyncer) QuotaSyncerService {
	return QuotaSyncerService{s: sync}
}

func (s QuotaSyncerService) Start() { s.s.StartService() }

func (s QuotaSyncerService) Stop() { s.s.Stop() }

// StatusPullerService 适配 scheduler.StatusPuller 为 service.Service。
type StatusPullerService struct {
	s *scheduler.StatusPuller
}

// NewStatusPullerService 创建短信状态主动查询扫描器适配器。
func NewStatusPullerService(p *scheduler.StatusPuller) StatusPullerService {
	return StatusPullerService{s: p}
}

func (s StatusPullerService) Start() { s.s.StartService() }

func (s StatusPullerService) Stop() { s.s.Stop() }
