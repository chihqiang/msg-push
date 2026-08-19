package sender

import (
	"encoding/json"
	"net/url"
	"testing"
)

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

func TestNeteaseFlexString(t *testing.T) {
	if got := neteaseFlexString("123"); got != "123" {
		t.Errorf("neteaseFlexString(string)=%q", got)
	}
	if got := neteaseFlexString(float64(123)); got != "123" {
		t.Errorf("neteaseFlexString(float)=%q, want 123", got)
	}
	if got := neteaseFlexString(nil); got != "" {
		t.Errorf("neteaseFlexString(nil)=%q, want empty", got)
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

func TestCanonicalizeAliyunParams(t *testing.T) {
	params := map[string]string{
		"Action":        "SendSms",
		"PhoneNum":      "138 0013 8000", // 含空格，验证 + → %20
		"Signature":     "a*b",           // 验证 * → %2A
		"Timestamp":     "2026-01-01T00:00:00Z",
		"TemplateParam": `{"code":"a b*c"}`,
	}
	got := canonicalizeAliyunParams(params)
	// 规范化后不得出现裸空格 / 未转义的 *（除 %2A）
	for i := 0; i < len(got); i++ {
		if got[i] == ' ' {
			t.Fatalf("canonicalize output contains raw space: %s", got)
		}
	}
	if !contains(got, "%20") {
		t.Errorf("expected %%20 encoding for space, got: %s", got)
	}
	if !contains(got, "%2A") {
		t.Errorf("expected %%2A encoding for *, got: %s", got)
	}
	// 关键：绝不能双重编码（%20 再次编码成 %2520）
	if contains(got, "%2520") || contains(got, "%252A") {
		t.Errorf("canonicalize output double-encoded: %s", got)
	}
}

// TestAliyunStringToSign 用阿里云官方文档示例验证 stringToSign 构造（修复双重编码 bug）。
// 官方示例：POST&%2F&AccessKeyId%3Dtestid%26Action%3DDescribeRegions%26Format%3DJSON%26...
func TestAliyunStringToSign(t *testing.T) {
	params := map[string]string{
		"AccessKeyId":      "testid",
		"Action":           "DescribeRegions",
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   "3ee8c1b8-83d3-44af-a94f-4e0ad82fd6cf",
		"SignatureVersion": "1.0",
		"Timestamp":        "2016-02-23T12:46:24Z",
		"Version":          "2014-05-26",
	}
	query := canonicalizeAliyunParams(params)
	// 一次编码后的规范化串（含 Timestamp 冒号编码）
	wantQuery := "AccessKeyId=testid&Action=DescribeRegions&Format=JSON&SignatureMethod=HMAC-SHA1" +
		"&SignatureNonce=3ee8c1b8-83d3-44af-a94f-4e0ad82fd6cf&SignatureVersion=1.0" +
		"&Timestamp=2016-02-23T12%3A46%3A24Z&Version=2014-05-26"
	if query != wantQuery {
		t.Fatalf("canonicalizeAliyunParams =\n  %s\nwant:\n  %s", query, wantQuery)
	}
	// stringToSign = POST&%2F& + url.QueryEscape(canonicalizedQueryString)
	stringToSign := "POST&%2F&" + url.QueryEscape(query)
	wantSign := "POST&%2F&AccessKeyId%3Dtestid%26Action%3DDescribeRegions%26Format%3DJSON" +
		"%26SignatureMethod%3DHMAC-SHA1%26SignatureNonce%3D3ee8c1b8-83d3-44af-a94f-4e0ad82fd6cf" +
		"%26SignatureVersion%3D1.0%26Timestamp%3D2016-02-23T12%253A46%253A24Z%26Version%3D2014-05-26"
	if stringToSign != wantSign {
		t.Fatalf("stringToSign =\n  %s\nwant:\n  %s", stringToSign, wantSign)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
