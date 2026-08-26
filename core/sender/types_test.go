package sender

import (
	"encoding/json"
	"testing"
	"time"

	"chihqiang/msg-push/model"
)

func TestBatchTaskParams(t *testing.T) {
	req := &BatchSendRequest{
		MappedParams: map[string]string{"k": "shared"},
		TaskParams:   []map[string]string{{"k": "t0"}, nil, {"k": "t2"}},
	}
	if got := batchTaskParams(req, 0)["k"]; got != "t0" {
		t.Errorf("batchTaskParams[0]=%q, want t0", got)
	}
	if got := batchTaskParams(req, 1)["k"]; got != "shared" {
		t.Errorf("batchTaskParams[1]=%q, want shared(回退)", got)
	}
	if got := batchTaskParams(req, 2)["k"]; got != "t2" {
		t.Errorf("batchTaskParams[2]=%q, want t2", got)
	}
	// 越界
	if got := batchTaskParams(req, 5); got == nil {
		t.Error("batchTaskParams(5) should fallback to MappedParams")
	}
}

func TestStatusQueryRequestFields(t *testing.T) {
	now := time.Now()
	req := &StatusQueryRequest{
		Task: &model.PushTask{TaskID: "t"}, ProviderAccount: newTestPA(nil),
		ProviderMsgID: "biz", PhoneNumber: "13800138000", SendDate: now,
	}
	if req.ProviderMsgID != "biz" || req.PhoneNumber != "13800138000" {
		t.Error("StatusQueryRequest fields not set")
	}
}

func TestTemplateKeys(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    []string
	}{
		{"无占位符", "hello world", nil},
		{"单占位符", "你好 {name}", []string{"name"}},
		{"多占位符按序", "验证码 {code}，有效期 {minute} 分钟", []string{"code", "minute"}},
		{"带下划线", "订单 {order_no} 已{status}", []string{"order_no", "status"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := templateKeys(c.content)
			if len(got) != len(c.want) {
				t.Fatalf("templateKeys(%q) = %v, want %v", c.content, got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("templateKeys(%q)[%d] = %q, want %q", c.content, i, got[i], c.want[i])
				}
			}
		})
	}
}

func TestStrValIntValBoolVal(t *testing.T) {
	cfg := map[string]any{
		"str":    "abc",
		"num_f":  float64(3.14),
		"num_i":  42,
		"flag_t": true,
		"flag_s": "true",
		"flag_1": "1",
	}
	if got := strVal(cfg, "str"); got != "abc" {
		t.Errorf("strVal(str)=%q", got)
	}
	if got := strVal(cfg, "num_f"); got != "3.14" {
		t.Errorf("strVal(num_f)=%q, want 3.14", got)
	}
	if got := strVal(cfg, "num_i"); got != "42" {
		t.Errorf("strVal(num_i)=%q, want 42", got)
	}
	if got := intVal(cfg, "num_f"); got != 3 {
		t.Errorf("intVal(num_f)=%d, want 3", got)
	}
	if got := boolVal(cfg, "flag_t"); !got {
		t.Errorf("boolVal(flag_t) should be true")
	}
	if got := boolVal(cfg, "flag_s"); !got {
		t.Errorf("boolVal(flag_s) should be true")
	}
	if got := boolVal(cfg, "flag_1"); !got {
		t.Errorf("boolVal(flag_1) should be true")
	}
	if got := boolVal(cfg, "missing"); got {
		t.Errorf("boolVal(missing) should be false")
	}
	if got := strVal(nil, "x"); got != "" {
		t.Errorf("strVal(nil)=%q, want empty", got)
	}
}

func TestIsAllDigits(t *testing.T) {
	if !isAllDigits("13800138000") {
		t.Error("isAllDigits(13800138000) should be true")
	}
	if isAllDigits("") {
		t.Error("isAllDigits('') should be false")
	}
	if isAllDigits("138a") {
		t.Error("isAllDigits(138a) should be false")
	}
}

func TestTruncateUTF8Bytes(t *testing.T) {
	s := "你好世界hello" // 4 个中文字符(12B) + 5 个 ASCII(5B) = 17B
	if got := truncateUTF8Bytes(s, 20); got != s {
		t.Errorf("truncate within limit = %q, want %q", got, s)
	}
	// 截断到 3 字节：恰好 "你"(3B)
	if got := truncateUTF8Bytes(s, 3); got != "你" {
		t.Errorf("truncate(3) = %q, want 你", got)
	}
	// 截断到 4 字节：第 4 字节是第二个汉字开头，应安全回退到 "你"
	if got := truncateUTF8Bytes(s, 4); got != "你" {
		t.Errorf("truncate(4) = %q, want 你（不切断多字节）", got)
	}
	// 截断到 5 字节：回退到 "你"（3B），不包含半个汉字
	if got := truncateUTF8Bytes(s, 5); got != "你" {
		t.Errorf("truncate(5) = %q, want 你（不切断多字节）", got)
	}
	// 负数/0 边界
	if got := truncateUTF8Bytes(s, 0); got != "" {
		t.Errorf("truncate(0) = %q, want empty", got)
	}
	// 纯 ASCII
	if got := truncateUTF8Bytes("abcdef", 3); got != "abc" {
		t.Errorf("truncate ascii = %q, want abc", got)
	}
}

func TestRawToAny(t *testing.T) {
	if got := rawToAny(json.RawMessage(`"abc"`)); got != "abc" {
		t.Errorf("rawToAny(string)=%v", got)
	}
	if got := rawToAny(json.RawMessage(`42`)); got != float64(42) {
		t.Errorf("rawToAny(number)=%v, want 42", got)
	}
	if got := rawToAny(nil); got != nil {
		t.Errorf("rawToAny(nil)=%v, want nil", got)
	}
}
