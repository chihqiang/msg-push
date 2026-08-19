package sender

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// 网易云信短信接口端点。
var (
	neteaseSendTemplateURL = "https://api.netease.im/sms/sendtemplate.action"
	neteaseSendCodeURL     = "https://api.netease.im/sms/sendcode.action"
)

// NeteaseSMSSender 网易云信短信发送器。
type NeteaseSMSSender struct{}

// GetProviderCode 返回服务商代码。
func (s *NeteaseSMSSender) GetProviderCode() string {
	return CodeNeteaseSMS
}

// neteaseFlexString 宽松解析字符串（兼容字符串/数字）。
func neteaseFlexString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%.0f", t)
	default:
		return ""
	}
}

// Send 发送短信。
func (s *NeteaseSMSSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
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
	appKey := strVal(cfg, "app_key")
	appSecret := strVal(cfg, "app_secret")
	sendType := strVal(cfg, "send_type")
	if sendType == "" {
		sendType = "template"
	}
	if appKey == "" || appSecret == "" {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "app_key/app_secret required", ErrorCode: "CONFIG_ERROR"}, nil
	}

	templateCode := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		templateCode = req.ChannelTemplateBinding.ProviderTemplate.TemplateCode
	}

	mobile := formatNeteaseMobile(req)
	nonce := randomHex(16)
	curTime := fmt.Sprintf("%d", time.Now().Unix())
	checkSum := hex.EncodeToString(sha1Sum([]byte(appSecret + nonce + curTime)))

	form := url.Values{}
	form.Set("templateid", templateCode)

	var sendURL string
	if sendType == "code" {
		// 验证码短信：单号接口，sendid 在 msg 字段
		sendURL = neteaseSendCodeURL
		form.Set("mobile", mobile)
		if len(req.MappedParams) > 0 {
			form.Set("paramMap", jsonDump(req.MappedParams))
		}
	} else {
		sendURL = neteaseSendTemplateURL
		form.Set("mobiles", jsonDump([]string{mobile}))
		// params 按模板占位符顺序
		params := orderedParams(req)
		form.Set("params", jsonDump(params))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, sendURL, strings.NewReader(form.Encode()))
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: form.Encode()}, nil
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	httpReq.Header.Set("AppKey", appKey)
	httpReq.Header.Set("Nonce", nonce)
	httpReq.Header.Set("CurTime", curTime)
	httpReq.Header.Set("CheckSum", checkSum)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: form.Encode()}, nil
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)
	responseData := jsonDump(map[string]any{"status": resp.StatusCode, "body": respBody.String()})

	var result struct {
		Code int             `json:"code"`
		Msg  json.RawMessage `json:"msg"`
		Obj  json.RawMessage `json:"obj"`
	}
	_ = json.Unmarshal([]byte(respBody.String()), &result)

	if result.Code == 200 {
		var sendID string
		if sendType == "code" {
			sendID = neteaseFlexString(rawToAny(result.Msg))
		} else {
			sendID = neteaseFlexString(rawToAny(result.Obj))
		}
		if sendID == "" {
			return &SendResponse{Success: false, TaskID: taskID, ErrorCode: "EMPTY_SENDID", ErrorMessage: "empty send id from netease", RequestData: form.Encode(), ResponseData: responseData}, nil
		}
		return &SendResponse{Success: true, TaskID: taskID, ProviderID: sendID, Status: string(model.PushTaskStatusSending), RequestData: form.Encode(), ResponseData: responseData}, nil
	}

	return &SendResponse{
		Success: false, TaskID: taskID, ErrorCode: fmt.Sprintf("%d", result.Code),
		ErrorMessage: string(result.Msg), RequestData: form.Encode(), ResponseData: responseData,
	}, nil
}

// formatNeteaseMobile 网易手机号格式化（复用通用 smsReceiver 规则）。
func formatNeteaseMobile(req *SendRequest) string {
	return smsReceiver(req)
}

// orderedParams 按模板占位符顺序取值（网易单条发送）。
func orderedParams(req *SendRequest) []string {
	content := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		content = req.ChannelTemplateBinding.ProviderTemplate.TemplateContent
	}
	return orderedParamsFromMap(req.MappedParams, content)
}

// SupportsBatchSend 是否支持批量发送。
func (s *NeteaseSMSSender) SupportsBatchSend() bool {
	return true
}

// BatchSend 批量发送短信（网易模板短信 mobiles 数组批量，单次最多 100 个号码，超出分批）。
// 验证码短信（send_type=code）仅支持单号接口，逐条发送。
// 批量成功时整批共用一个真实 sendid；回执按 (provider_msg_id + receiver) 关联到具体任务。
func (s *NeteaseSMSSender) BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error) {
	if len(req.Tasks) == 0 {
		return &BatchSendResponse{Results: []*SendResponse{}}, nil
	}
	if req.ProviderAccount == nil {
		return &BatchSendResponse{Results: s.neteaseFailAll(req.Tasks, "provider account missing", "CONFIG_ERROR")}, nil
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return &BatchSendResponse{Results: s.neteaseFailAll(req.Tasks, "invalid config: "+err.Error(), "CONFIG_ERROR")}, nil
	}
	appKey := strVal(cfg, "app_key")
	appSecret := strVal(cfg, "app_secret")
	sendType := strVal(cfg, "send_type")
	if sendType == "" {
		sendType = "template"
	}
	if appKey == "" || appSecret == "" {
		return &BatchSendResponse{Results: s.neteaseFailAll(req.Tasks, "app_key/app_secret required", "CONFIG_ERROR")}, nil
	}

	templateCode := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		templateCode = req.ChannelTemplateBinding.ProviderTemplate.TemplateCode
	}
	if templateCode == "" {
		return &BatchSendResponse{Results: s.neteaseFailAll(req.Tasks, "missing template_code", "CONFIG_ERROR")}, nil
	}

	// 验证码短信：单号接口，逐条发送
	if sendType == "code" {
		results := make([]*SendResponse, len(req.Tasks))
		for i, task := range req.Tasks {
			results[i] = s.neteaseSendCodeOne(ctx, appKey, appSecret, templateCode, task, batchTaskParams(req, i))
		}
		return &BatchSendResponse{Results: results}, nil
	}

	// 模板短信：mobiles 数组批量，最多 100/批，超出分批
	shared := req.MappedParams
	if len(req.TaskParams) > 0 && req.TaskParams[0] != nil {
		shared = req.TaskParams[0]
	}
	content := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		content = req.ChannelTemplateBinding.ProviderTemplate.TemplateContent
	}
	params := orderedParamsFromMap(shared, content)

	results := make([]*SendResponse, 0, len(req.Tasks))
	const maxBatch = 100
	for start := 0; start < len(req.Tasks); start += maxBatch {
		end := start + maxBatch
		if end > len(req.Tasks) {
			end = len(req.Tasks)
		}
		chunk := req.Tasks[start:end]
		results = append(results, s.neteaseSendTemplateChunk(ctx, appKey, appSecret, templateCode, params, chunk, req)...)
	}
	return &BatchSendResponse{Results: results}, nil
}

// neteaseSendCodeOne 网易验证码单号发送。
func (s *NeteaseSMSSender) neteaseSendCodeOne(ctx context.Context, appKey, appSecret, templateCode string, task *model.PushTask, mapped map[string]string) *SendResponse {
	mobile := task.Receiver
	nonce := randomHex(16)
	curTime := fmt.Sprintf("%d", time.Now().Unix())
	checkSum := hex.EncodeToString(sha1Sum([]byte(appSecret + nonce + curTime)))

	form := url.Values{}
	form.Set("templateid", templateCode)
	form.Set("mobile", mobile)
	if len(mapped) > 0 {
		form.Set("paramMap", jsonDump(mapped))
	}

	requestData := form.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, neteaseSendCodeURL, strings.NewReader(requestData))
	if err != nil {
		return &SendResponse{Success: false, TaskID: task.TaskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: requestData}
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	httpReq.Header.Set("AppKey", appKey)
	httpReq.Header.Set("Nonce", nonce)
	httpReq.Header.Set("CurTime", curTime)
	httpReq.Header.Set("CheckSum", checkSum)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &SendResponse{Success: false, TaskID: task.TaskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: requestData}
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)
	responseData := jsonDump(map[string]any{"status": resp.StatusCode, "body": respBody.String()})

	var result struct {
		Code int             `json:"code"`
		Msg  json.RawMessage `json:"msg"`
	}
	_ = json.Unmarshal([]byte(respBody.String()), &result)
	if result.Code == 200 {
		sendID := neteaseFlexString(rawToAny(result.Msg))
		if sendID == "" {
			return &SendResponse{Success: false, TaskID: task.TaskID, ErrorCode: "EMPTY_SENDID", ErrorMessage: "empty send id from netease", RequestData: requestData, ResponseData: responseData}
		}
		return &SendResponse{Success: true, TaskID: task.TaskID, ProviderID: sendID, Status: string(model.PushTaskStatusSending), RequestData: requestData, ResponseData: responseData}
	}
	return &SendResponse{
		Success: false, TaskID: task.TaskID, ErrorCode: fmt.Sprintf("%d", result.Code),
		ErrorMessage: string(result.Msg), RequestData: requestData, ResponseData: responseData,
	}
}

// neteaseSendTemplateChunk 发送一个不超过 100 个手机号的模板短信批次。
func (s *NeteaseSMSSender) neteaseSendTemplateChunk(ctx context.Context, appKey, appSecret, templateCode string, params []string, tasks []*model.PushTask, req *BatchSendRequest) []*SendResponse {
	mobiles := make([]string, len(tasks))
	for i, task := range tasks {
		mobiles[i] = task.Receiver
	}

	nonce := randomHex(16)
	curTime := fmt.Sprintf("%d", time.Now().Unix())
	checkSum := hex.EncodeToString(sha1Sum([]byte(appSecret + nonce + curTime)))

	form := url.Values{}
	form.Set("templateid", templateCode)
	form.Set("mobiles", jsonDump(mobiles))
	form.Set("params", jsonDump(params))

	requestData := form.Encode()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, neteaseSendTemplateURL, strings.NewReader(requestData))
	if err != nil {
		return s.neteaseFailAll(tasks, err.Error(), "HTTP_ERROR")
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	httpReq.Header.Set("AppKey", appKey)
	httpReq.Header.Set("Nonce", nonce)
	httpReq.Header.Set("CurTime", curTime)
	httpReq.Header.Set("CheckSum", checkSum)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return s.neteaseFailAll(tasks, err.Error(), "HTTP_ERROR")
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)
	responseData := jsonDump(map[string]any{"status": resp.StatusCode, "body": respBody.String()})

	var result struct {
		Code int             `json:"code"`
		Obj  json.RawMessage `json:"obj"`
		Msg  json.RawMessage `json:"msg"`
	}
	_ = json.Unmarshal([]byte(respBody.String()), &result)

	results := make([]*SendResponse, len(tasks))
	if result.Code == 200 {
		sendID := neteaseFlexString(rawToAny(result.Obj))
		if sendID == "" {
			return s.neteaseFailAll(tasks, "empty send id from netease", "EMPTY_SENDID")
		}
		for i, task := range tasks {
			results[i] = &SendResponse{
				Success: true, TaskID: task.TaskID, ProviderID: sendID,
				Status: string(model.PushTaskStatusSending), RequestData: requestData, ResponseData: responseData,
			}
		}
		return results
	}
	for i, task := range tasks {
		results[i] = &SendResponse{
			Success: false, TaskID: task.TaskID, ErrorCode: fmt.Sprintf("%d", result.Code),
			ErrorMessage: string(result.Msg), RequestData: requestData, ResponseData: responseData,
		}
	}
	return results
}

// orderedParamsFromMap 按模板占位符顺序取值（网易批量共用参数）。
// 网易 params 为数组，顺序必须与模板占位符一致；无模板占位符时按 key 排序稳定回退。
func orderedParamsFromMap(mapped map[string]string, content string) []string {
	if len(mapped) == 0 {
		return nil
	}
	keys := templateKeys(content)
	if len(keys) == 0 {
		return sortedParamsFromMap(mapped)
	}
	vals := make([]string, 0, len(keys))
	for _, k := range keys {
		vals = append(vals, mapped[k])
	}
	return vals
}

// neteaseFailAll 为批量失败场景生成全部失败结果。
func (s *NeteaseSMSSender) neteaseFailAll(tasks []*model.PushTask, errMsg, errCode string) []*SendResponse {
	results := make([]*SendResponse, len(tasks))
	for i, task := range tasks {
		results[i] = &SendResponse{Success: false, TaskID: task.TaskID, ErrorCode: errCode, ErrorMessage: errMsg}
	}
	return results
}
