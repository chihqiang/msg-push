package sender

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// 阿里云短信接口端点。
var aliyunSmsEndpoint = "https://dysmsapi.aliyuncs.com/"

// AliyunSMSSender 阿里云短信发送器（RPC V1 签名，标准库实现）。
type AliyunSMSSender struct{}

// GetProviderCode 返回服务商代码。
func (s *AliyunSMSSender) GetProviderCode() string {
	return CodeAliyunSMS
}

// Send 发送短信。
func (s *AliyunSMSSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
	taskID := ""
	if req.Task != nil {
		taskID = req.Task.TaskID
	}
	receiver := smsReceiver(req)
	if req.ProviderAccount == nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "provider account missing", ErrorCode: "CONFIG_ERROR"}, nil
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "invalid config: " + err.Error(), ErrorCode: "CONFIG_ERROR"}, nil
	}
	accessKeyID := strVal(cfg, "access_key_id")
	accessKeySecret := strVal(cfg, "access_key_secret")
	if accessKeyID == "" || accessKeySecret == "" {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "access_key_id/access_key_secret required", ErrorCode: "CONFIG_ERROR"}, nil
	}

	signName := ""
	if req.Signature != nil {
		signName = req.Signature.SignatureCode
	}
	templateCode := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		templateCode = req.ChannelTemplateBinding.ProviderTemplate.TemplateCode
	}
	templateParam := ""
	if len(req.MappedParams) > 0 {
		templateParam = jsonDump(req.MappedParams)
	}

	// 公共 + 业务参数
	params := map[string]string{
		"Action":           "SendSms",
		"Version":          "2017-05-25",
		"Format":           "JSON",
		"AccessKeyId":      accessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   randomHex(16),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"RegionId":         "cn-hangzhou",
		"PhoneNumbers":     receiver,
		"SignName":         signName,
		"TemplateCode":     templateCode,
	}
	if templateParam != "" {
		params["TemplateParam"] = templateParam
	}

	// 规范化查询串
	query := canonicalizeAliyunParams(params)

	// 签名
	stringToSign := "POST&%2F&" + url.QueryEscape(query)
	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	form.Set("Signature", signature)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, aliyunSmsEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: form.Encode()}, nil
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		Code    string `json:"Code"`
		Message string `json:"Message"`
		BizId   string `json:"BizId"`
	}
	_ = json.Unmarshal([]byte(respBody.String()), &result)

	if result.Code == "OK" {
		return &SendResponse{Success: true, TaskID: taskID, ProviderID: result.BizId, Status: string(model.PushTaskStatusSending), RequestData: form.Encode(), ResponseData: responseData}, nil
	}
	return &SendResponse{
		Success: false, TaskID: taskID, ErrorCode: result.Code,
		ErrorMessage: result.Message, RequestData: form.Encode(), ResponseData: responseData,
	}, nil
}

// canonicalizeAliyunParams 阿里云参数规范化：key 排序 + RFC3986 编码特判，返回一次编码后的查询串。
// 注意：仅做一次编码。签名时由调用方用 url.QueryEscape 整体二次编码（阿里云 stringToSign 要求）。
func canonicalizeAliyunParams(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	encoded := values.Encode() // 内部按 key 排序
	// 从 application/x-www-form-urlencoded 特判转换为 RFC3986 编码
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

// QueryStatus 查询短信发送状态（阿里云 QuerySendDetails API）。
// 阿里云状态：SendStatus 1-等待回执(视为 unknown)，2-发送失败，3-发送成功。
func (s *AliyunSMSSender) QueryStatus(ctx context.Context, req *StatusQueryRequest) (*StatusQueryResponse, error) {
	if req == nil || req.ProviderAccount == nil {
		return nil, fmt.Errorf("invalid status query request")
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}
	accessKeyID := strVal(cfg, "access_key_id")
	accessKeySecret := strVal(cfg, "access_key_secret")
	if accessKeyID == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("access_key_id/access_key_secret required")
	}

	params := map[string]string{
		"Action":           "QuerySendDetails",
		"Version":          "2017-05-25",
		"Format":           "JSON",
		"AccessKeyId":      accessKeyID,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   randomHex(16),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"RegionId":         "cn-hangzhou",
		"PhoneNumber":      req.PhoneNumber,
		"SendDate":         req.SendDate.Format("20060102"),
		"PageSize":         "10",
		"CurrentPage":      "1",
	}
	if req.ProviderMsgID != "" {
		params["BizId"] = req.ProviderMsgID
	}

	query := canonicalizeAliyunParams(params)
	stringToSign := "POST&%2F&" + url.QueryEscape(query)
	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	form.Set("Signature", signature)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, aliyunSmsEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)

	var result struct {
		Code              string `json:"Code"`
		Message           string `json:"Message"`
		SmsSendDetailDTOs struct {
			SmsSendDetailDTO []struct {
				PhoneNum   string `json:"PhoneNum"`
				SendStatus int64  `json:"SendStatus"`
				ErrCode    string `json:"ErrCode"`
			} `json:"SmsSendDetailDTO"`
		} `json:"SmsSendDetailDTOs"`
	}
	if err := json.Unmarshal([]byte(respBody.String()), &result); err != nil {
		return nil, fmt.Errorf("unmarshal query response: %w", err)
	}
	if result.Code != "OK" {
		return nil, fmt.Errorf("query failed: code=%s, message=%s", result.Code, result.Message)
	}

	out := &StatusQueryResponse{Results: make([]*StatusQueryResult, 0, len(result.SmsSendDetailDTOs.SmsSendDetailDTO))}
	for _, d := range result.SmsSendDetailDTOs.SmsSendDetailDTO {
		status := "unknown"
		switch d.SendStatus {
		case 2:
			status = "failed"
		case 3:
			status = "delivered"
		}
		out.Results = append(out.Results, &StatusQueryResult{
			ProviderMsgID: req.ProviderMsgID,
			PhoneNumber:   d.PhoneNum,
			Status:        status,
			ErrorCode:     d.ErrCode,
		})
	}
	return out, nil
}

// SupportsBatchSend 是否支持批量发送。
func (s *AliyunSMSSender) SupportsBatchSend() bool {
	return true
}

// BatchSend 批量发送短信（阿里云 SendBatchSms API，单次最多 1000 个号码）。
// 批量成功时服务商仅返回一个统一 BizId；每个号码的最终回执通过回调按
// (provider_msg_id + receiver) 尽力关联到具体任务。
func (s *AliyunSMSSender) BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error) {
	if len(req.Tasks) == 0 {
		return &BatchSendResponse{Results: []*SendResponse{}}, nil
	}
	if req.ProviderAccount == nil {
		return &BatchSendResponse{Results: s.failAll(req.Tasks, "provider account missing", "CONFIG_ERROR")}, nil
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return &BatchSendResponse{Results: s.failAll(req.Tasks, "invalid config: "+err.Error(), "CONFIG_ERROR")}, nil
	}
	accessKeyID := strVal(cfg, "access_key_id")
	accessKeySecret := strVal(cfg, "access_key_secret")
	if accessKeyID == "" || accessKeySecret == "" {
		return &BatchSendResponse{Results: s.failAll(req.Tasks, "access_key_id/access_key_secret required", "CONFIG_ERROR")}, nil
	}

	signName := ""
	if req.Signature != nil {
		signName = req.Signature.SignatureCode
	}
	templateCode := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		templateCode = req.ChannelTemplateBinding.ProviderTemplate.TemplateCode
	}
	if templateCode == "" {
		return &BatchSendResponse{Results: s.failAll(req.Tasks, "missing template_code", "CONFIG_ERROR")}, nil
	}

	// 构造逐号码 JSON 数组
	phoneNumbers := make([]string, len(req.Tasks))
	signNames := make([]string, len(req.Tasks))
	templateParams := make([]string, len(req.Tasks))
	for i, task := range req.Tasks {
		phoneNumbers[i] = task.Receiver
		signNames[i] = signName
		p := batchTaskParams(req, i)
		if len(p) > 0 {
			templateParams[i] = jsonDump(p)
		} else {
			templateParams[i] = "{}"
		}
	}
	phoneNumbersJSON, _ := json.Marshal(phoneNumbers)
	signNamesJSON, _ := json.Marshal(signNames)
	templateParamsJSON, _ := json.Marshal(templateParams)

	params := map[string]string{
		"Action":            "SendBatchSms",
		"Version":           "2017-05-25",
		"Format":            "JSON",
		"AccessKeyId":       accessKeyID,
		"SignatureMethod":   "HMAC-SHA1",
		"SignatureVersion":  "1.0",
		"SignatureNonce":    randomHex(16),
		"Timestamp":         time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"RegionId":          "cn-hangzhou",
		"PhoneNumberJson":   string(phoneNumbersJSON),
		"SignNameJson":      string(signNamesJSON),
		"TemplateCode":      templateCode,
		"TemplateParamJson": string(templateParamsJSON),
	}

	query := canonicalizeAliyunParams(params)
	stringToSign := "POST&%2F&" + url.QueryEscape(query)
	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}
	form.Set("Signature", signature)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, aliyunSmsEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return &BatchSendResponse{Results: s.failAll(req.Tasks, err.Error(), "HTTP_ERROR")}, nil
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &BatchSendResponse{Results: s.failAll(req.Tasks, err.Error(), "HTTP_ERROR")}, nil
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)
	responseData := jsonDump(map[string]any{"status": resp.StatusCode, "body": respBody.String()})

	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
		BizId   string `json:"BizId"`
	}
	_ = json.Unmarshal([]byte(respBody.String()), &result)

	results := make([]*SendResponse, len(req.Tasks))
	if result.Code == "OK" {
		for i, task := range req.Tasks {
			results[i] = &SendResponse{
				Success: true, TaskID: task.TaskID, ProviderID: result.BizId,
				Status: string(model.PushTaskStatusSending), RequestData: form.Encode(), ResponseData: responseData,
			}
		}
		return &BatchSendResponse{Results: results}, nil
	}
	for i, task := range req.Tasks {
		results[i] = &SendResponse{
			Success: false, TaskID: task.TaskID, ErrorCode: result.Code,
			ErrorMessage: result.Message, RequestData: form.Encode(), ResponseData: responseData,
		}
	}
	return &BatchSendResponse{Results: results}, nil
}

// failAll 为批量失败场景生成全部失败结果。
func (s *AliyunSMSSender) failAll(tasks []*model.PushTask, errMsg, errCode string) []*SendResponse {
	results := make([]*SendResponse, len(tasks))
	for i, task := range tasks {
		results[i] = &SendResponse{Success: false, TaskID: task.TaskID, ErrorCode: errCode, ErrorMessage: errMsg}
	}
	return results
}

// ==================== 阿里云短信回执解析 ====================

// AliyunCallbackParser 阿里云短信回执解析器。
// 格式：JSON 数组 [{"phone_number":"138****","send_time":"...","report_time":"...","success":true,"err_code":"DELIVRD","err_msg":"...","biz_id":"12345^67890",...}]
type AliyunCallbackParser struct{}

// GetProviderCode 返回服务商代码。
func (p *AliyunCallbackParser) GetProviderCode() string {
	return CodeAliyunSMS
}

// Parse 解析阿里云回执。
func (p *AliyunCallbackParser) Parse(ctx context.Context, req *CallbackRequest) (CallbackResponse, []*CallbackResult, error) {
	var reports []struct {
		PhoneNumber string `json:"phone_number"`
		ReportTime  string `json:"report_time"`
		Success     bool   `json:"success"`
		ErrCode     string `json:"err_code"`
		ErrMsg      string `json:"err_msg"`
		BizID       string `json:"biz_id"`
	}
	if err := json.Unmarshal(req.RawBody, &reports); err != nil {
		// 解析失败也返回成功响应，避免服务商重复推送
		return AliyunCallbackOK, nil, fmt.Errorf("invalid aliyun callback: %w", err)
	}

	results := make([]*CallbackResult, 0, len(reports))
	for _, r := range reports {
		status := model.CallbackStatusDelivered
		if !r.Success {
			status = model.CallbackStatusFailed
		}
		reportTime, _ := time.ParseInLocation("2006-01-02 15:04:05", r.ReportTime, time.Local)
		results = append(results, &CallbackResult{
			Type:         string(model.CallbackTypeReport),
			ProviderID:   r.BizID,
			Status:       status,
			ErrorCode:    r.ErrCode,
			ErrorMessage: r.ErrMsg,
			Mobile:       r.PhoneNumber,
			ReportTime:   reportTime,
		})
	}
	return AliyunCallbackOK, results, nil
}
