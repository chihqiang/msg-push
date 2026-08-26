package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chihqiang/msg-push/model"
)

// DingTalkSender 钉钉工作通知发送器。
type DingTalkSender struct{}

// GetProviderCode 返回服务商代码。
func (s *DingTalkSender) GetProviderCode() string {
	return CodeDingTalk
}

// dingTalkTokenResponse token 响应。
type dingTalkTokenResponse struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// getDingTalkToken 获取钉钉 access_token（带缓存）。
func getDingTalkToken(ctx context.Context, appKey, appSecret string) (string, error) {
	key := "dingtalk:" + appKey
	if token, ok := getCachedToken(key); ok {
		return token, nil
	}
	url := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s", appKey, appSecret)
	body, _, err := httpGet(ctx, url, 10*time.Second)
	if err != nil {
		return "", err
	}
	var resp dingTalkTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if resp.Errcode != 0 {
		return "", fmt.Errorf("dingtalk gettoken failed: %s", resp.Errmsg)
	}
	ttl := time.Duration(resp.ExpiresIn-200) * time.Second
	if ttl <= 0 {
		ttl = 7200 * time.Second
	}
	setCachedToken(key, resp.AccessToken, ttl)
	return resp.AccessToken, nil
}

// Send 发送钉钉工作通知。
func (s *DingTalkSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
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
	appKey := strVal(cfg, "app_key")
	appSecret := strVal(cfg, "app_secret")
	agentID := intVal(cfg, "agent_id")
	if appKey == "" || appSecret == "" || agentID == 0 {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "app_key/app_secret/agent_id required", ErrorCode: "CONFIG_ERROR"}, nil
	}

	token, err := getDingTalkToken(ctx, appKey, appSecret)
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "TOKEN_ERROR"}, nil
	}

	msgtype := "text"
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil &&
		req.ChannelTemplateBinding.ProviderTemplate.ContentType == "markdown" {
		msgtype = "markdown"
	}

	var msgContent map[string]any
	if msgtype == "markdown" {
		title := "消息"
		if req.Task != nil && req.Task.Signature != "" {
			title = req.Task.Signature
		}
		msgContent = map[string]any{"msgtype": "markdown", "markdown": map[string]any{"title": title, "text": req.RenderedContent}}
	} else {
		msgContent = map[string]any{"msgtype": "text", "text": map[string]any{"content": truncateUTF8Bytes(req.RenderedContent, 2048)}}
	}

	payload := map[string]any{
		"agent_id":    agentID,
		"userid_list": receiver,
		"msg":         msgContent,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://oapi.dingtalk.com/topapi/message/corpconversation/asyncsend_v2?access_token=%s", token)
	respBody, statusCode, err := httpPost(ctx, url, body, 10*time.Second)
	responseData := jsonDump(map[string]any{"status": statusCode, "body": string(respBody)})
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: string(body), ResponseData: responseData}, nil
	}

	var resp struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		TaskID  int64  `json:"task_id"`
	}
	_ = json.Unmarshal(respBody, &resp)
	if resp.Errcode == 0 {
		return &SendResponse{Success: true, TaskID: taskID, ProviderID: fmt.Sprintf("%d", resp.TaskID), Status: string(model.PushTaskStatusSuccess), RequestData: string(body), ResponseData: responseData}, nil
	}
	return &SendResponse{
		Success: false, TaskID: taskID, ErrorCode: fmt.Sprintf("%d", resp.Errcode),
		ErrorMessage: resp.Errmsg, RequestData: string(body), ResponseData: responseData,
	}, nil
}
