// Package common 提供 core 各子包共享的公共工具（HTTP/JSON/字符串/退避等），
// 消除 pipeline / scheduler / sender 之间的重复实现。
package common

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

// HTTPTransport 共享 HTTP Transport（连接池），供各发送器复用。
// 复用 TCP 连接减少握手开销；整体超时由每次请求的 ctx 或 client.Timeout 控制。
var HTTPTransport = &http.Transport{
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 16,
	IdleConnTimeout:     90 * time.Second,
	TLSHandshakeTimeout: 10 * time.Second,
	DialContext:         (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
}

// PostJSON HTTP POST 辅助（JSON 请求体），复用共享连接池。
// 返回响应体与状态码；整体超时由 ctx 或 timeout 控制。
func PostJSON(ctx context.Context, url string, body []byte, timeout time.Duration) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeout, Transport: HTTPTransport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, resp.Body)
	return buf.Bytes(), resp.StatusCode, nil
}

// Get HTTP GET 辅助，复用共享连接池。
// 返回响应体与状态码；整体超时由 ctx 或 timeout 控制。
func Get(ctx context.Context, url string, timeout time.Duration) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	client := &http.Client{Timeout: timeout, Transport: HTTPTransport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, resp.Body)
	return buf.Bytes(), resp.StatusCode, nil
}
