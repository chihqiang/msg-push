package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// WeChatWorkRobotSender 企业微信群机器人发送器。
type WeChatWorkRobotSender struct{}

// GetProviderCode 返回服务商代码。
func (s *WeChatWorkRobotSender) GetProviderCode() string {
	return CodeWeChatWorkRobot
}

// Send 发送群机器人消息。
func (s *WeChatWorkRobotSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
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
	msgType := strVal(cfg, "msg_type")
	if msgType == "" {
		msgType = "text"
	}

	var payload map[string]any
	switch msgType {
	case "markdown":
		content := req.RenderedContent
		if req.Task != nil {
			content = buildWeChatRobotMarkdownMentions(content, req.Task.Receiver)
		}
		payload = map[string]any{
			"msgtype":  "markdown",
			"markdown": map[string]any{"content": content},
		}
	default:
		content := truncateUTF8Bytes(req.RenderedContent, 4096)
		mentionedList, mentionedMobileList := parseWeChatRobotMentions(req.Task)
		payload = map[string]any{
			"msgtype": "text",
			"text": map[string]any{
				"content":               content,
				"mentioned_list":        mentionedList,
				"mentioned_mobile_list": mentionedMobileList,
			},
		}
	}

	body, _ := json.Marshal(payload)
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

// parseWeChatRobotMentions 解析企微机器人 @ 列表。
func parseWeChatRobotMentions(task *model.PushTask) (mentionedList, mentionedMobileList []string) {
	if task == nil || task.Receiver == "" {
		return []string{}, []string{}
	}
	for _, part := range strings.Split(task.Receiver, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if p == "@all" || p == "all" {
			mentionedList = append(mentionedList, "@all")
		} else if isAllDigits(p) {
			mentionedMobileList = append(mentionedMobileList, p)
		} else {
			mentionedList = append(mentionedList, p)
		}
	}
	return mentionedList, mentionedMobileList
}

// buildWeChatRobotMarkdownMentions markdown 的 @ 前缀。
func buildWeChatRobotMarkdownMentions(content, receiver string) string {
	if content == "" {
		content = "消息"
	}
	var mentions []string
	for _, part := range strings.Split(receiver, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		if p == "@all" || p == "all" {
			mentions = append(mentions, "<@all>")
		} else if !isAllDigits(p) {
			mentions = append(mentions, "<@"+p+">")
		}
	}
	if len(mentions) > 0 {
		return strings.Join(mentions, " ") + "\n" + content
	}
	return content
}
