package servicex

import (
	"chihqiang/msg-push/scheduler"
)

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
