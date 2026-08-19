package middleware

import (
	"testing"
)

func TestGenerateAndVerifySignature(t *testing.T) {
	secret := "demo-secret"
	params := map[string]any{"name": "张三", "code": "123456"}

	sig := generateSignature(secret, "POST", "/api/v1/messages", params, 1786735818, "nonce-001")
	if sig == "" {
		t.Fatal("empty signature")
	}

	// 正确签名应通过
	if !verifySignature(secret, "POST", "/api/v1/messages", params, 1786735818, "nonce-001", sig) {
		t.Error("valid signature should verify")
	}

	// 篡改任一输入后重新签名，验证应失败
	badCases := []struct {
		name  string
		sigFn func() string
	}{
		{"secret", func() string {
			return generateSignature("other", "POST", "/api/v1/messages", params, 1786735818, "nonce-001")
		}},
		{"method", func() string {
			return generateSignature(secret, "GET", "/api/v1/messages", params, 1786735818, "nonce-001")
		}},
		{"path", func() string {
			return generateSignature(secret, "POST", "/api/v1/other", params, 1786735818, "nonce-001")
		}},
		{"params", func() string {
			return generateSignature(secret, "POST", "/api/v1/messages", map[string]any{"name": "李四"}, 1786735818, "nonce-001")
		}},
		{"timestamp", func() string {
			return generateSignature(secret, "POST", "/api/v1/messages", params, 1786735819, "nonce-001")
		}},
		{"nonce", func() string {
			return generateSignature(secret, "POST", "/api/v1/messages", params, 1786735818, "nonce-002")
		}},
	}
	for _, c := range badCases {
		t.Run(c.name, func(t *testing.T) {
			badSig := c.sigFn()
			if badSig == sig {
				t.Fatal("tamper case did not change the signature (test ineffective)")
			}
			if verifySignature(secret, "POST", "/api/v1/messages", params, 1786735818, "nonce-001", badSig) {
				t.Errorf("signature with different %s should NOT verify", c.name)
			}
		})
	}

	// 直接断言：错误签名验证失败
	wrong := generateSignature("other-secret", "POST", "/api/v1/messages", params, 1786735818, "nonce-001")
	if verifySignature(secret, "POST", "/api/v1/messages", params, 1786735818, "nonce-001", wrong) {
		t.Error("wrong-secret signature should not verify")
	}
}

func TestSortedJSONParams(t *testing.T) {
	// key 排序后 JSON 序列化，值类型保持
	got := sortedJSONParams(map[string]any{"b": 2, "a": 1, "c": "x"})
	want := `{"a":1,"b":2,"c":"x"}`
	if got != want {
		t.Errorf("sortedJSONParams = %s, want %s", got, want)
	}
	// 空参数
	if got := sortedJSONParams(nil); got != "" {
		t.Errorf("empty params = %q, want empty", got)
	}
}

func TestHMACSHA256KnownVector(t *testing.T) {
	// 已知向量：HMAC-SHA256(key="key", data="The quick brown fox jumps over the lazy dog")
	got := hmacSHA256("The quick brown fox jumps over the lazy dog", "key")
	want := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if got != want {
		t.Errorf("hmacSHA256 = %s, want %s", got, want)
	}
}

func TestVerifySignatureConstantTime(t *testing.T) {
	// 长度不同也安全返回 false（不 panic）
	if verifySignature("sec", "POST", "/p", nil, 1786735818, "n", "abc") {
		t.Error("should not verify short signature")
	}
}
