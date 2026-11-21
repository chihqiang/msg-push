package sender

import (
	"testing"
)

func TestCallbackParserRegistry(t *testing.T) {
	for _, code := range []string{"aliyun_sms", "tencent_sms", "netease_sms"} {
		if p := GetCallbackParser(code); p == nil {
			t.Errorf("callback parser not registered for %s", code)
		}
	}
	if p := GetCallbackParser("smtp"); p != nil {
		t.Errorf("smtp should not have a callback parser")
	}
}

func TestNormalizeCallbackStatus(t *testing.T) {
	cases := map[string]string{
		"DELIVRD":      "delivered",
		"delivered":    "delivered",
		"SUCCESS":      "delivered",
		"sent":         "delivered",
		"failed":       "failed",
		"undelivered":  "failed",
		"rejected":     "rejected",
		"refuse":       "rejected",
		"some_unknown": "unknown",
		"":             "unknown",
	}
	for in, want := range cases {
		if got := normalizeCallbackStatus(in); got != want {
			t.Errorf("normalizeCallbackStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
