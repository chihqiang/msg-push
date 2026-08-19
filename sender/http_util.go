package sender

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

// httpPost HTTP POST 辅助（JSON 请求体）。
func httpPost(ctx context.Context, url string, body []byte, timeout time.Duration) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	return buf.Bytes(), resp.StatusCode, nil
}

// copyBuffer 从 reader 拷贝到 builder。
func copyBuffer(w io.Writer, r io.Reader) (int64, error) {
	return io.Copy(w, r)
}
