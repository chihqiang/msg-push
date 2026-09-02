package sender

import (
	"context"
	"time"

	"chihqiang/msg-push/core/common"
)

// ==================== common 工具兼容短名 ====================
// 以下为 core/common 公共工具在 sender 包内的兼容短名（实现统一在 core/common），
// 供各发送器内部沿用简短调用，避免大范围改动既有调用点。

// senderTransport 共享 HTTP Transport（连接池），复用 core/common 连接池。
var senderTransport = common.HTTPTransport

// httpPost HTTP POST 辅助（JSON 请求体）。
func httpPost(ctx context.Context, url string, body []byte, timeout time.Duration) ([]byte, int, error) {
	return common.PostJSON(ctx, url, body, timeout)
}

// httpGet HTTP GET 辅助。
func httpGet(ctx context.Context, url string, timeout time.Duration) ([]byte, int, error) {
	return common.Get(ctx, url, timeout)
}

// jsonDump 序列化调试数据。
func jsonDump(v any) string {
	return common.MarshalJSONString(v)
}

// randomHex 生成 n 字节随机数的 hex 字符串。
func randomHex(n int) string {
	return common.RandomHex(n)
}

// truncateUTF8Bytes 按 UTF-8 字节安全截断（不切断多字节字符）。
func truncateUTF8Bytes(s string, maxBytes int) string {
	return common.TruncateUTF8(s, maxBytes)
}

// isAllDigits 是否全数字。
func isAllDigits(s string) bool {
	return common.IsAllDigits(s)
}
