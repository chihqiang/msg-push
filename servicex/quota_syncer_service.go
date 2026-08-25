package servicex

import (
	"chihqiang/msg-push/scheduler"
)

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
