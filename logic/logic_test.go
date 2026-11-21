package logic

import (
	"testing"
	"time"

	"chihqiang/msg-push/core/sender"
	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
)

func TestMarshalParams(t *testing.T) {
	got, err := marshalParams(map[string]string{"name": "张三"})
	if err != nil || got != `{"name":"张三"}` {
		t.Errorf("marshalParams = %q, err=%v", got, err)
	}
	// 空参数返回空串
	if got, _ := marshalParams(nil); got != "" {
		t.Errorf("nil params -> %q, want empty", got)
	}
}

func TestMarshalHelpers(t *testing.T) {
	got, err := marshalStringSlice([]string{"a", "b"})
	if err != nil || got != `["a","b"]` {
		t.Errorf("marshalStringSlice = %q, err=%v", got, err)
	}
	mapping := []model.ParamMappingItem{{Type: "mapping", SystemVar: "name", ProviderVar: "n"}}
	got, err = marshalParamMapping(mapping)
	if err != nil || got != `[{"type":"mapping","provider_var":"n","system_var":"name","value":""}]` {
		t.Errorf("marshalParamMapping = %q, err=%v", got, err)
	}
}

func TestParseCallbackJSON(t *testing.T) {
	pc := parseCallback([]byte(`{"message_id":"m1","status":"delivered","err_code":"","mobile":"13800138000"}`), nil, nil)
	if pc.ProviderID != "m1" || pc.Status != "delivered" || pc.Mobile != "13800138000" {
		t.Errorf("parseCallback = %+v", pc)
	}

	// 数组 receipts 形态
	pc = parseCallback([]byte(`{"receipts":[{"bizId":"b2","status":"failed","err_code":"E1"}]}`), nil, nil)
	if pc.ProviderID != "b2" || pc.Status != "failed" || pc.ErrorCode != "E1" {
		t.Errorf("parseCallback array = %+v", pc)
	}
}

func TestParseCallbackFormFallback(t *testing.T) {
	form := map[string]string{"sid": "s1", "report_status": "SUCCESS"}
	pc := parseCallback(nil, form, nil)
	if pc.ProviderID != "s1" {
		t.Errorf("form provider id = %q", pc.ProviderID)
	}
	if pc.Status != "delivered" {
		t.Errorf("form status = %q, want delivered", pc.Status)
	}
}

func TestNormalizeStatus(t *testing.T) {
	cases := map[string]string{
		"":            "",
		"delivered":   "delivered",
		"DELIVERED":   "delivered",
		"success":     "delivered",
		"sent":        "delivered",
		"failed":      "failed",
		"fail":        "failed",
		"undelivered": "failed",
		"rejected":    "rejected",
		"refuse":      "rejected",
		"something":   "unknown",
	}
	for in, want := range cases {
		if got := normalizeStatus(in); got != want {
			t.Errorf("normalizeStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstStrAndNonEmpty(t *testing.T) {
	m := map[string]any{"a": "va", "b": float64(42), "c": ""}
	if got := firstStr(m, "x", "a"); got != "va" {
		t.Errorf("firstStr = %q", got)
	}
	if got := firstStr(m, "x", "b"); got != "42" {
		t.Errorf("firstStr float = %q, want 42", got)
	}
	if got := firstStr(m, "x", "c"); got != "" {
		t.Errorf("firstStr empty should skip, got %q", got)
	}
	if got := firstNonEmpty("k", map[string]string{"k": "v1"}, map[string]string{"k": "v2"}); got != "v1" {
		t.Errorf("firstNonEmpty = %q, want v1", got)
	}
	if got := firstNonEmpty("k", map[string]string{"j": "x"}); got != "" {
		t.Errorf("firstNonEmpty missing = %q", got)
	}
}

func TestQuotaKey(t *testing.T) {
	now := time.Date(2026, 8, 15, 10, 0, 0, 0, time.Local)
	if got := quotaKey(7, now); got != "quota:7:20260815" {
		t.Errorf("quotaKey = %q", got)
	}
}

func TestStatisticsDateBucket(t *testing.T) {
	cases := map[string]string{
		"mysql":    "DATE_FORMAT(created_at, '%Y-%m-%d')",
		"postgres": "TO_CHAR(created_at, 'YYYY-MM-DD')",
		"sqlite":   "strftime('%Y-%m-%d', created_at)",
		"other":    "DATE(created_at)",
	}
	for in, want := range cases {
		if got := statisticsDateBucketExpression(in); got != want {
			t.Errorf("statisticsDateBucketExpression(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveRange(t *testing.T) {
	now := time.Date(2026, 8, 15, 10, 0, 0, 0, time.Local)

	// 默认近 30 天
	r, err := resolveRange(&dto.StatisticsRequest{}, now)
	if err != nil {
		t.Fatalf("default range: %v", err)
	}
	if !r.start.Equal(time.Date(2026, 7, 17, 0, 0, 0, 0, time.Local)) {
		t.Errorf("default start = %v, want 2026-07-17", r.start)
	}
	if !r.end.Equal(time.Date(2026, 8, 16, 0, 0, 0, 0, time.Local)) {
		t.Errorf("default end(exclusive) = %v, want 2026-08-16", r.end)
	}

	// 指定范围
	r, err = resolveRange(&dto.StatisticsRequest{StartDate: "2026-08-01", EndDate: "2026-08-15"}, now)
	if err != nil {
		t.Fatalf("explicit range: %v", err)
	}
	if !r.start.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)) {
		t.Errorf("explicit start = %v", r.start)
	}

	// 只给一个日期报错
	if _, err := resolveRange(&dto.StatisticsRequest{StartDate: "2026-08-01"}, now); err == nil {
		t.Error("single date should error")
	}
	// 非法日期报错
	if _, err := resolveRange(&dto.StatisticsRequest{StartDate: "bad", EndDate: "2026-08-15"}, now); err == nil {
		t.Error("invalid date should error")
	}
	// 超 90 天报错
	if _, err := resolveRange(&dto.StatisticsRequest{StartDate: "2026-01-01", EndDate: "2026-08-15"}, now); err == nil {
		t.Error(">90 days should error")
	}
	// nil 请求报错
	if _, err := resolveRange(nil, now); err == nil {
		t.Error("nil request should error")
	}
}

func TestBusinessDayStart(t *testing.T) {
	now := time.Date(2026, 8, 15, 23, 59, 59, 0, time.Local)
	got := businessDayStart(now)
	if got.Hour() != 0 || got.Minute() != 0 || got.Day() != 15 {
		t.Errorf("businessDayStart = %v", got)
	}
}

func TestCompletedSuccessRate(t *testing.T) {
	if r := completedSuccessRate(0, 0); r != nil {
		t.Errorf("no completion should be nil, got %v", *r)
	}
	if r := completedSuccessRate(3, 1); r == nil || *r != 75 {
		t.Errorf("3/4 rate = %v, want 75", r)
	}
	if r := completedSuccessRate(1, 0); r == nil || *r != 100 {
		t.Errorf("1/1 rate = %v, want 100", r)
	}
}

func TestLegacySuccessRate(t *testing.T) {
	if got := legacySuccessRate(0, 0); got != "0.00%" {
		t.Errorf("legacySuccessRate(0,0) = %q", got)
	}
	if got := legacySuccessRate(1, 4); got != "25.00%" {
		t.Errorf("legacySuccessRate(1,4) = %q", got)
	}
}

func TestValidateProviderConfig(t *testing.T) {
	meta := &sender.Meta{
		ConfigFields: []sender.ConfigField{
			{Key: "access_key_id", Label: "AccessKeyId", Required: true},
			{Key: "access_key_secret", Label: "AccessKeySecret", Required: true},
			{Key: "region", Label: "区域", Required: false},
		},
	}
	// 缺失必填
	err := validateProviderConfig(meta, map[string]any{"access_key_id": "x"})
	if err == nil {
		t.Error("missing required field should error")
	}
	// 全部必填满足
	err = validateProviderConfig(meta, map[string]any{"access_key_id": "x", "access_key_secret": "y"})
	if err != nil {
		t.Errorf("valid config should pass, got %v", err)
	}
	// 空字符串视为缺失
	err = validateProviderConfig(meta, map[string]any{"access_key_id": "", "access_key_secret": "y"})
	if err == nil {
		t.Error("empty required value should error")
	}
}

func TestMarshalProviderConfig(t *testing.T) {
	got, err := marshalProviderConfig(map[string]any{"host": "smtp.x.com"})
	if err != nil || got != `{"host":"smtp.x.com"}` {
		t.Errorf("marshalProviderConfig = %q, err=%v", got, err)
	}
}
