package sender

import (
	"chihqiang/msg-push/model"
)

// newTestPA 构造带配置的测试服务商账号。
func newTestPA(cfg map[string]any) *model.ProviderAccount {
	pa := &model.ProviderAccount{}
	_ = pa.SetConfig(cfg)
	return pa
}

// newSMSRequest 构造短信发送请求。
func newSMSRequest(pa *model.ProviderAccount, receiver string) *SendRequest {
	return &SendRequest{
		Task:            &model.PushTask{TaskID: "task_t1", Receiver: receiver},
		ProviderAccount: pa,
		ChannelTemplateBinding: &model.ChannelTemplateBinding{
			ProviderTemplate: &model.ProviderTemplate{TemplateCode: "TPL_001", TemplateContent: "你好 {name}"},
		},
		Signature:    &model.ProviderSignature{SignatureCode: "SIG_ALI"},
		MappedParams: map[string]string{"name": "张三"},
	}
}

// hexStr 字节转小写 hex 字符串。
func hexStr(b []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = digits[v>>4]
		out[i*2+1] = digits[v&0xf]
	}
	return string(out)
}

// contains 判断字符串是否包含子串。
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
