package sender

import (
	"regexp"
	"unicode/utf8"
)

// cnMobileRe 中国大陆手机号正则。
var cnMobileRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// isAllDigits 是否全数字。
func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// truncateUTF8Bytes 按 UTF-8 字节安全截断（不切断多字节字符）。
func truncateUTF8Bytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	b := []byte(s)
	cut := b[:maxBytes]
	// 回退到最后一个完整 UTF-8 字符边界
	for len(cut) > 0 && !utf8.Valid(cut) {
		cut = cut[:len(cut)-1]
	}
	return string(cut)
}
