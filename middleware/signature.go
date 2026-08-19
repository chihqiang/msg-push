package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// hmacSHA256 计算 HMAC-SHA256，返回 hex 编码字符串。
func hmacSHA256(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// sortedJSONParams 将请求体参数按 key 排序后 JSON 序列化，用于签名计算。
// 签名规范：signature = HMAC(secret, method+path+sortedParams+timestamp+nonce)。
func sortedJSONParams(params map[string]any) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sorted := make(map[string]any, len(params))
	for _, k := range keys {
		sorted[k] = params[k]
	}
	b, _ := json.Marshal(sorted)
	return string(b)
}

// generateSignature 生成 HMAC-SHA256 签名。
func generateSignature(secret, method, path string, params map[string]any, timestamp int64, nonce string) string {
	signContent := fmt.Sprintf("%s%s%s%d%s", method, path, sortedJSONParams(params), timestamp, nonce)
	return hmacSHA256(signContent, secret)
}

// verifySignature 校验签名，使用常量时间比较防时序攻击。
func verifySignature(secret, method, path string, params map[string]any, timestamp int64, nonce, signature string) bool {
	expected := generateSignature(secret, method, path, params, timestamp, nonce)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}
