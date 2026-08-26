package sender

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// DingTalkRobotSender 钉钉群机器人发送器。
type DingTalkRobotSender struct{}

// GetProviderCode 返回服务商代码。
func (s *DingTalkRobotSender) GetProviderCode() string {
	return CodeDingTalkRobot
}

// Send 发送钉钉群机器人消息（支持加签）。
func (s *DingTalkRobotSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
	taskID := ""
	if req.Task != nil {
		taskID = req.Task.TaskID
	}
	if req.ProviderAccount == nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "provider account missing", ErrorCode: "CONFIG_ERROR"}, nil
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "invalid config: " + err.Error(), ErrorCode: "CONFIG_ERROR"}, nil
	}
	webhookURL := strVal(cfg, "webhook_url")
	if webhookURL == "" {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "webhook_url required", ErrorCode: "CONFIG_ERROR"}, nil
	}
	secret := strVal(cfg, "secret")
	msgType := strVal(cfg, "msg_type")
	if msgType == "" {
		msgType = "text"
	}

	atMobiles, atUserIDs, isAtAll := parseDingTalkRobotAt(req.Task)

	var payload map[string]any
	switch msgType {
	case "markdown":
		title := "通知"
		if req.Task != nil && req.Task.Signature != "" {
			title = req.Task.Signature
		}
		payload = map[string]any{
			"msgtype":  "markdown",
			"markdown": map[string]any{"title": title, "text": req.RenderedContent},
			"at":       map[string]any{"isAtAll": isAtAll, "atMobiles": atMobiles, "atUserIds": atUserIDs},
		}
	default:
		payload = map[string]any{
			"msgtype": "text",
			"text":    map[string]any{"content": truncateUTF8Bytes(req.RenderedContent, 4096)},
			"at":      map[string]any{"isAtAll": isAtAll, "atMobiles": atMobiles, "atUserIds": atUserIDs},
		}
	}

	body, _ := json.Marshal(payload)

	// 加签
	if secret != "" {
		ts := time.Now().UnixMilli()
		sign := dingTalkRobotSign(secret, ts)
		sep := "?"
		if strings.Contains(webhookURL, "?") {
			sep = "&"
		}
		webhookURL = fmt.Sprintf("%s%stimestamp=%d&sign=%s", webhookURL, sep, ts, url.QueryEscape(sign))
	}

	respBody, statusCode, err := httpPost(ctx, webhookURL, body, 10*time.Second)
	responseData := jsonDump(map[string]any{"status": statusCode, "body": string(respBody)})
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorCode: "HTTP_ERROR", ErrorMessage: err.Error(), RequestData: string(body), ResponseData: responseData}, nil
	}

	var resp struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	_ = json.Unmarshal(respBody, &resp)
	if resp.Errcode == 0 {
		return &SendResponse{Success: true, TaskID: taskID, Status: string(model.PushTaskStatusSuccess), RequestData: string(body), ResponseData: responseData}, nil
	}
	return &SendResponse{
		Success: false, TaskID: taskID, ErrorCode: fmt.Sprintf("%d", resp.Errcode),
		ErrorMessage: resp.Errmsg, RequestData: string(body), ResponseData: responseData,
	}, nil
}

// dingTalkRobotSign 钉钉机器人加签。
func dingTalkRobotSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// parseDingTalkRobotAt 解析钉钉机器人 @ 列表。
func parseDingTalkRobotAt(task *model.PushTask) (atMobiles, atUserIDs []string, isAtAll bool) {
	if task == nil || task.Receiver == "" {
		return []string{}, []string{}, false
	}
	for _, part := range strings.Split(task.Receiver, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if p == "@all" || p == "all" {
			isAtAll = true
		} else if isAllDigits(p) {
			atMobiles = append(atMobiles, p)
		} else {
			atUserIDs = append(atUserIDs, p)
		}
	}
	return atMobiles, atUserIDs, isAtAll
}
