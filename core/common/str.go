package common

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"unicode/utf8"
)

// CNMobileRe 中国大陆手机号正则。
var CNMobileRe = regexp.MustCompile(`^1[3-9]\d{9}$`)

// IsAllDigits 是否全为数字字符。
func IsAllDigits(s string) bool {
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

// TruncateUTF8 按 UTF-8 字节安全截断（不切断多字节字符）。
func TruncateUTF8(s string, maxBytes int) string {
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

// RandomHex 生成 n 字节随机数的 hex 字符串。
func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// MarshalJSON 序列化（失败返回 nil）。
func MarshalJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// MarshalJSONString 序列化调试数据（失败返回 "{}"）。
func MarshalJSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// UnmarshalJSONString 反序列化字符串为 JSON；空串直接返回 nil。
func UnmarshalJSONString(data string, dst any) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), dst)
}
