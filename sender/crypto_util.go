package sender

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
)

// randomHex 生成 n 字节随机数的 hex 字符串。
func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// sha1Sum 计算 SHA1 摘要。
func sha1Sum(b []byte) []byte {
	h := sha1.New()
	_, _ = h.Write(b)
	return h.Sum(nil)
}
