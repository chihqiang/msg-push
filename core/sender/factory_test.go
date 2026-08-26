package sender

import (
	"testing"
)

func TestFactoryRegisteredSenders(t *testing.T) {
	f := NewFactory()
	for _, code := range []string{
		CodeAliyunSMS, CodeTencentSMS, CodeNeteaseSMS, CodeSMTP,
		CodeWeChatWork, CodeDingTalk, CodeWeChatWorkRobot, CodeDingTalkRobot,
	} {
		if _, err := f.GetSender(code); err != nil {
			t.Errorf("sender not registered for %s: %v", code, err)
		}
	}
	if _, err := f.GetSender("nonexistent"); err == nil {
		t.Error("GetSender(nonexistent) should error")
	}
}

func TestFactoryBatchSender(t *testing.T) {
	f := NewFactory()
	// 三个短信服务商支持批量
	for _, code := range []string{CodeAliyunSMS, CodeTencentSMS, CodeNeteaseSMS} {
		if _, err := f.GetBatchSender(code); err != nil {
			t.Errorf("batch sender not available for %s: %v", code, err)
		}
	}
	// SMTP 不支持批量
	if _, err := f.GetBatchSender(CodeSMTP); err == nil {
		t.Error("GetBatchSender(smtp) should error (not supported)")
	}
}

func TestDefaultResolver(t *testing.T) {
	s, err := DefaultResolver.GetSender(CodeAliyunSMS)
	if err != nil || s == nil {
		t.Fatalf("DefaultResolver.GetSender(aliyun_sms) failed: %v", err)
	}
	if s.GetProviderCode() != CodeAliyunSMS {
		t.Errorf("sender code = %s, want %s", s.GetProviderCode(), CodeAliyunSMS)
	}
}
