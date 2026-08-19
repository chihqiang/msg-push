package worker

import (
	"testing"

	"chihqiang/msg-push/model"
)

func TestMatchRuleBasics(t *testing.T) {
	e := &ruleEngine{} // 仅测 matchRule 纯逻辑，无需 DB

	req := &EvaluateRequest{
		Scene:        model.RuleSceneSendFailure,
		ProviderCode: "aliyun_sms",
		MessageType:  "sms",
		ErrorCode:    "ISV_SMS_SIGNATURE_ILLEGAL",
		ErrorMessage: "签名不合法 signature illegal",
		Task:         &model.PushTask{},
	}

	cases := []struct {
		name string
		rule *model.FailureRule
		want bool
	}{
		{"全空条件通配", &model.FailureRule{}, true},
		{"provider 匹配", &model.FailureRule{ProviderCode: "aliyun_sms"}, true},
		{"provider 不匹配", &model.FailureRule{ProviderCode: "tencent_sms"}, false},
		{"message_type 匹配", &model.FailureRule{MessageType: "sms"}, true},
		{"message_type 不匹配", &model.FailureRule{MessageType: "email"}, false},
		{"error_code 单值匹配", &model.FailureRule{ErrorCode: "ISV_SMS_SIGNATURE_ILLEGAL"}, true},
		{"error_code 多值命中", &model.FailureRule{ErrorCode: "A,ISV_SMS_SIGNATURE_ILLEGAL,B"}, true},
		{"error_code 不匹配", &model.FailureRule{ErrorCode: "OTHER"}, false},
		{"error_keyword 模糊命中(大小写不敏感)", &model.FailureRule{ErrorKeyword: "SIGNATURE"}, true},
		{"error_keyword 多关键字", &model.FailureRule{ErrorKeyword: "foo,签名"}, true},
		{"error_keyword 不命中", &model.FailureRule{ErrorKeyword: "不存在的词"}, false},
		{"多条件 AND 全满足", &model.FailureRule{ProviderCode: "aliyun_sms", MessageType: "sms", ErrorCode: "ISV_SMS_SIGNATURE_ILLEGAL"}, true},
		{"多条件 AND 有一不满足", &model.FailureRule{ProviderCode: "aliyun_sms", MessageType: "sms", ErrorCode: "OTHER"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := e.matchRule(c.rule, req); got != c.want {
				t.Errorf("matchRule = %v, want %v", got, c.want)
			}
		})
	}
}

func TestDefaultResult(t *testing.T) {
	e := &ruleEngine{}
	if got := e.defaultResult(model.RuleSceneSendFailure).Action; got != model.RuleActionRetry {
		t.Errorf("send_failure default = %s, want retry", got)
	}
	if got := e.defaultResult(model.RuleSceneCallbackFailure).Action; got != model.RuleActionFail {
		t.Errorf("callback_failure default = %s, want fail", got)
	}
	if got := e.defaultResult("unknown").Action; got != model.RuleActionFail {
		t.Errorf("unknown default = %s, want fail", got)
	}
}
