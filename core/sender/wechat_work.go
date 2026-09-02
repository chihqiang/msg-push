package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"chihqiang/msg-push/model"
)

// tokenCacheItem token 缓存项。
type tokenCacheItem struct {
	token     string
	expiresAt time.Time
}

// tokenCache 进程内 token 缓存（key → item）。
var tokenCache = struct {
	sync.Mutex
	m map[string]tokenCacheItem
}{m: map[string]tokenCacheItem{}}

// getCachedToken 获取缓存 token（未过期返回）。
func getCachedToken(key string) (string, bool) {
	tokenCache.Lock()
	defer tokenCache.Unlock()
	item, ok := tokenCache.m[key]
	if !ok {
		return "", false
	}
	if time.Now().After(item.expiresAt) {
		delete(tokenCache.m, key)
		return "", false
	}
	return item.token, true
}

// setCachedToken 缓存 token。
func setCachedToken(key, token string, ttl time.Duration) {
	tokenCache.Lock()
	defer tokenCache.Unlock()
	tokenCache.m[key] = tokenCacheItem{token: token, expiresAt: time.Now().Add(ttl)}
}

// WeChatWorkSender 企业微信应用消息发送器。
type WeChatWorkSender struct{}

// GetProviderCode 返回服务商代码。
func (s *WeChatWorkSender) GetProviderCode() string {
	return CodeWeChatWork
}

// wechatWorkTokenResponse token 响应。
type wechatWorkTokenResponse struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// getWechatWorkToken 获取企业微信 access_token（带缓存）。
func getWechatWorkToken(ctx context.Context, corpID, agentSecret string) (string, error) {
	key := "wechat_work:" + corpID + ":" + agentSecret
	if token, ok := getCachedToken(key); ok {
		return token, nil
	}
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, agentSecret)
	body, _, err := httpGet(ctx, url, 10*time.Second)
	if err != nil {
		return "", err
	}
	var resp wechatWorkTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if resp.Errcode != 0 {
		return "", fmt.Errorf("wechat work gettoken failed: %s", resp.Errmsg)
	}
	ttl := time.Duration(resp.ExpiresIn-200) * time.Second
	if ttl <= 0 {
		ttl = 7200 * time.Second
	}
	setCachedToken(key, resp.AccessToken, ttl)
	return resp.AccessToken, nil
}

// Send 发送企业微信应用消息。
func (s *WeChatWorkSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
	taskID := ""
	receiver := ""
	if req.Task != nil {
		taskID = req.Task.TaskID
		receiver = req.Task.Receiver
	}
	if req.ProviderAccount == nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "provider account missing", ErrorCode: "CONFIG_ERROR"}, nil
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "invalid config: " + err.Error(), ErrorCode: "CONFIG_ERROR"}, nil
	}
	corpID := strVal(cfg, "corp_id")
	agentSecret := strVal(cfg, "agent_secret")
	agentID := intVal(cfg, "agent_id")
	if corpID == "" || agentSecret == "" || agentID == 0 {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "corp_id/agent_secret/agent_id required", ErrorCode: "CONFIG_ERROR"}, nil
	}

	token, err := getWechatWorkToken(ctx, corpID, agentSecret)
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "TOKEN_ERROR"}, nil
	}

	touser := receiver
	if touser == "" {
		touser = "@all"
	}
	content := truncateUTF8Bytes(req.RenderedContent, 2048)

	msgtype := "text"
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil &&
		req.ChannelTemplateBinding.ProviderTemplate.ContentType == "markdown" {
		msgtype = "markdown"
	}

	payload := map[string]any{
		"touser":  touser,
		"msgtype": msgtype,
		"agentid": agentID,
		"safe":    0,
	}
	if msgtype == "markdown" {
		payload["markdown"] = map[string]any{"content": content}
	} else {
		payload["text"] = map[string]any{"content": content}
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	respBody, statusCode, err := httpPost(ctx, url, body, 10*time.Second)
	responseData := jsonDump(map[string]any{"status": statusCode, "body": string(respBody)})
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: string(body), ResponseData: responseData}, nil
	}

	var resp struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		Msgid   string `json:"msgid"`
	}
	_ = json.Unmarshal(respBody, &resp)
	if resp.Errcode == 0 {
		return &SendResponse{Success: true, TaskID: taskID, ProviderID: resp.Msgid, Status: string(model.PushTaskStatusSuccess), RequestData: string(body), ResponseData: responseData}, nil
	}
	return &SendResponse{
		Success: false, TaskID: taskID, ErrorCode: fmt.Sprintf("%d", resp.Errcode),
		ErrorMessage: resp.Errmsg, RequestData: string(body), ResponseData: responseData,
	}, nil
}

// httpGet HTTP GET 辅助。
func httpGet(ctx context.Context, url string, timeout time.Duration) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	client := &http.Client{Timeout: timeout, Transport: senderTransport}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	buf := new(strings.Builder)
	_, _ = copyBuffer(buf, resp.Body)
	return []byte(buf.String()), resp.StatusCode, nil
}
