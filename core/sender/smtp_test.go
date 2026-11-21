package sender

import (
	"testing"
)

func TestNormalizeEmailSubject(t *testing.T) {
	if got := normalizeEmailSubject(""); got != "通知" {
		t.Errorf("empty -> %q, want 通知", got)
	}
	if got := normalizeEmailSubject("  验证码  "); got != "验证码" {
		t.Errorf("trim -> %q", got)
	}
	if got := normalizeEmailSubject("标题\r\n注入"); got != "通知" {
		t.Errorf("with CRLF should fallback, got %q", got)
	}
	if got := normalizeEmailSubject("正常标题"); got != "正常标题" {
		t.Errorf("normal -> %q", got)
	}
}
