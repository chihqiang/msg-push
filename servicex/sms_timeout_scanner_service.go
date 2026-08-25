package servicex

import (
	"chihqiang/msg-push/scheduler"
)

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
