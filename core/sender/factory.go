package sender

import (
	"fmt"
)

// ==================== 发送器工厂 ====================

// Factory 发送器工厂，按 providerCode 注册 8 个发送器。
type Factory struct {
	senders map[string]Sender
}

// NewFactory 创建并注册全部发送器。
func NewFactory() *Factory {
	f := &Factory{senders: map[string]Sender{}}
	f.Register(&AliyunSMSSender{})
	f.Register(&TencentSMSSender{})
	f.Register(&NeteaseSMSSender{})
	f.Register(&SMTPSender{})
	f.Register(&WeChatWorkSender{})
	f.Register(&DingTalkSender{})
	f.Register(&WeChatWorkRobotSender{})
	f.Register(&DingTalkRobotSender{})
	return f
}

// Register 注册发送器。
func (f *Factory) Register(s Sender) {
	f.senders[s.GetProviderCode()] = s
}

// GetSender 根据服务商代码获取发送器。
func (f *Factory) GetSender(providerCode string) (Sender, error) {
	s, ok := f.senders[providerCode]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerCode)
	}
	return s, nil
}

// GetBatchSender 根据服务商代码获取批量发送器；不支持时返回错误。
func (f *Factory) GetBatchSender(providerCode string) (BatchSender, error) {
	s, err := f.GetSender(providerCode)
	if err != nil {
		return nil, err
	}
	b, ok := s.(BatchSender)
	if !ok || !b.SupportsBatchSend() {
		return nil, fmt.Errorf("provider %s does not support batch send", providerCode)
	}
	return b, nil
}

// DefaultResolver 全局默认解析器（由 SetDefaultResolver 注入）。
var DefaultResolver Resolver = NewFactory()

// SetDefaultResolver 设置全局解析器。
func SetDefaultResolver(r Resolver) {
	if r != nil {
		DefaultResolver = r
	}
}
