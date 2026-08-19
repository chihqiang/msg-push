package sender

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// 腾讯云短信接口端点与版本。
var tencentSmsEndpoint = "https://sms.tencentcloudapi.com/"

const tencentSmsVersion = "2021-01-11"

// TencentSMSSender 腾讯云短信发送器（TC3-HMAC-SHA256 签名，标准库实现）。
type TencentSMSSender struct{}

// GetProviderCode 返回服务商代码。
func (s *TencentSMSSender) GetProviderCode() string {
	return CodeTencentSMS
}

// Send 发送短信。
func (s *TencentSMSSender) Send(ctx context.Context, req *SendRequest) (*SendResponse, error) {
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
	secretID := strVal(cfg, "secret_id")
	secretKey := strVal(cfg, "secret_key")
	sdkAppID := strVal(cfg, "sdk_app_id")
	if secretID == "" || secretKey == "" || sdkAppID == "" {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: "secret_id/secret_key/sdk_app_id required", ErrorCode: "CONFIG_ERROR"}, nil
	}
	region := strVal(cfg, "region")
	if region == "" {
		region = "ap-guangzhou"
	}

	signName := ""
	if req.Signature != nil {
		signName = req.Signature.SignatureCode
	}
	templateID := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		templateID = req.ChannelTemplateBinding.ProviderTemplate.TemplateCode
	}
	templateParamSet := tencentBuildParams(req)

	payload := map[string]any{
		"SmsSdkAppId":      sdkAppID,
		"SignName":         signName,
		"TemplateId":       templateID,
		"PhoneNumberSet":   []string{receiver},
		"TemplateParamSet": templateParamSet,
	}
	payloadBytes, _ := json.Marshal(payload)

	// TC3 签名
	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	authorization := tencentTC3Sign(secretID, secretKey, date, timestamp, payloadBytes)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tencentSmsEndpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: string(payloadBytes)}, nil
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Host", "sms.tencentcloudapi.com")
	httpReq.Header.Set("X-TC-Action", "SendSms")
	httpReq.Header.Set("X-TC-Version", tencentSmsVersion)
	httpReq.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	httpReq.Header.Set("X-TC-Region", region)
	httpReq.Header.Set("Authorization", authorization)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &SendResponse{Success: false, TaskID: taskID, ErrorMessage: err.Error(), ErrorCode: "HTTP_ERROR", RequestData: string(payloadBytes)}, nil
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)
	responseData := jsonDump(map[string]any{"status": resp.StatusCode, "body": respBody.String()})

	var result struct {
		Response struct {
			SendStatusSet []struct {
				SerialNo string `json:"SerialNo"`
				Code     string `json:"Code"`
				Message  string `json:"Message"`
			} `json:"SendStatusSet"`
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	_ = json.Unmarshal([]byte(respBody.String()), &result)

	if result.Response.Error.Code != "" {
		return &SendResponse{
			Success: false, TaskID: taskID, ErrorCode: result.Response.Error.Code,
			ErrorMessage: result.Response.Error.Message, RequestData: string(payloadBytes), ResponseData: responseData,
		}, nil
	}
	if len(result.Response.SendStatusSet) > 0 {
		item := result.Response.SendStatusSet[0]
		if item.Code == "Ok" {
			return &SendResponse{Success: true, TaskID: taskID, ProviderID: item.SerialNo, Status: string(model.PushTaskStatusSending), RequestData: string(payloadBytes), ResponseData: responseData}, nil
		}
		return &SendResponse{
			Success: false, TaskID: taskID, ErrorCode: item.Code,
			ErrorMessage: item.Message, RequestData: string(payloadBytes), ResponseData: responseData,
		}, nil
	}
	return &SendResponse{
		Success: false, TaskID: taskID, ErrorCode: "EMPTY_RESPONSE",
		ErrorMessage: "empty send status from tencent", RequestData: string(payloadBytes), ResponseData: responseData,
	}, nil
}

// tencentBuildParams 腾讯云模板参数按占位符顺序（单条发送）。
func tencentBuildParams(req *SendRequest) []string {
	content := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		content = req.ChannelTemplateBinding.ProviderTemplate.TemplateContent
	}
	return tencentBuildParamsFromMap(req.MappedParams, content)
}

// QueryStatus 查询短信发送状态（腾讯云 PullSmsSendStatusByPhoneNumber API）。
func (s *TencentSMSSender) QueryStatus(ctx context.Context, req *StatusQueryRequest) (*StatusQueryResponse, error) {
	if req == nil || req.ProviderAccount == nil {
		return nil, fmt.Errorf("invalid status query request")
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return nil, fmt.Errorf("invalid provider config: %w", err)
	}
	secretID := strVal(cfg, "secret_id")
	secretKey := strVal(cfg, "secret_key")
	sdkAppID := strVal(cfg, "sdk_app_id")
	if secretID == "" || secretKey == "" || sdkAppID == "" {
		return nil, fmt.Errorf("secret_id/secret_key/sdk_app_id required")
	}
	region := strVal(cfg, "region")
	if region == "" {
		region = "ap-guangzhou"
	}

	payload := map[string]any{
		"SmsSdkAppId": sdkAppID,
		"PhoneNumber": req.PhoneNumber,
		"BeginTime":   req.SendDate.Unix(),
		"EndTime":     req.SendDate.Add(24 * time.Hour).Unix(),
		"Offset":      0,
		"Limit":       100,
	}
	payloadBytes, _ := json.Marshal(payload)

	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	authorization := tencentTC3Sign(secretID, secretKey, date, timestamp, payloadBytes)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tencentSmsEndpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Host", "sms.tencentcloudapi.com")
	httpReq.Header.Set("X-TC-Action", "PullSmsSendStatusByPhoneNumber")
	httpReq.Header.Set("X-TC-Version", tencentSmsVersion)
	httpReq.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	httpReq.Header.Set("X-TC-Region", region)
	httpReq.Header.Set("Authorization", authorization)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)

	var result struct {
		Response struct {
			PullSmsSendStatusSet []struct {
				SerialNo     string `json:"SerialNo"`
				PhoneNumber  string `json:"PhoneNumber"`
				ReportStatus string `json:"ReportStatus"`
				Description  string `json:"Description"`
			} `json:"PullSmsSendStatusSet"`
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.Unmarshal([]byte(respBody.String()), &result); err != nil {
		return nil, fmt.Errorf("unmarshal query response: %w", err)
	}
	if result.Response.Error.Code != "" {
		return nil, fmt.Errorf("query failed: code=%s, message=%s", result.Response.Error.Code, result.Response.Error.Message)
	}

	out := &StatusQueryResponse{Results: make([]*StatusQueryResult, 0, len(result.Response.PullSmsSendStatusSet))}
	for _, d := range result.Response.PullSmsSendStatusSet {
		status := "unknown"
		if d.ReportStatus == "SUCCESS" {
			status = "delivered"
		} else if d.ReportStatus == "FAIL" {
			status = "failed"
		}
		out.Results = append(out.Results, &StatusQueryResult{
			ProviderMsgID: d.SerialNo,
			PhoneNumber:   d.PhoneNumber,
			Status:        status,
			ErrorCode:     d.ReportStatus,
			ErrorMessage:  d.Description,
		})
	}
	return out, nil
}

// tencentTC3Sign 计算腾讯云 TC3-HMAC-SHA256 签名。
func tencentTC3Sign(secretID, secretKey, date string, timestamp int64, payload []byte) string {
	// CanonicalRequest
	canonicalHeaders := "content-type:application/json; charset=utf-8\nhost:sms.tencentcloudapi.com\n"
	signedHeaders := "content-type;host"
	hashedPayload := sha256Hex(payload)
	canonicalRequest := fmt.Sprintf("POST\n/\n\n%s\n%s\n%s", canonicalHeaders, signedHeaders, hashedPayload)

	// StringToSign
	hashedCanonicalRequest := sha256Hex([]byte(canonicalRequest))
	stringToSign := fmt.Sprintf("TC3-HMAC-SHA256\n%d\n%s/sms/tc3_request\n%s", timestamp, date, hashedCanonicalRequest)

	// 派生密钥
	secretDate := hmacSHA256([]byte("TC3"+secretKey), []byte(date))
	secretService := hmacSHA256(secretDate, []byte("sms"))
	secretSigning := hmacSHA256(secretService, []byte("tc3_request"))

	// Signature
	signature := hex.EncodeToString(hmacSHA256(secretSigning, []byte(stringToSign)))

	// Authorization
	return fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s/sms/tc3_request, SignedHeaders=%s, Signature=%s",
		secretID, date, signedHeaders, signature)
}

// sha256Hex SHA256 hex。
func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// hmacSHA256 HMAC-SHA256。
func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// SupportsBatchSend 是否支持批量发送。
func (s *TencentSMSSender) SupportsBatchSend() bool {
	return true
}

// BatchSend 批量发送短信（腾讯云 SendSms 原生支持多号码，单次最多 200 个）。
// 服务商在 SendStatusSet 中逐号码返回结果，按顺序映射回各任务。
func (s *TencentSMSSender) BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error) {
	if len(req.Tasks) == 0 {
		return &BatchSendResponse{Results: []*SendResponse{}}, nil
	}
	if req.ProviderAccount == nil {
		return &BatchSendResponse{Results: s.tencentFailAll(req.Tasks, "provider account missing", "CONFIG_ERROR")}, nil
	}
	cfg, err := configMap(req.ProviderAccount)
	if err != nil {
		return &BatchSendResponse{Results: s.tencentFailAll(req.Tasks, "invalid config: "+err.Error(), "CONFIG_ERROR")}, nil
	}
	secretID := strVal(cfg, "secret_id")
	secretKey := strVal(cfg, "secret_key")
	sdkAppID := strVal(cfg, "sdk_app_id")
	if secretID == "" || secretKey == "" || sdkAppID == "" {
		return &BatchSendResponse{Results: s.tencentFailAll(req.Tasks, "secret_id/secret_key/sdk_app_id required", "CONFIG_ERROR")}, nil
	}
	region := strVal(cfg, "region")
	if region == "" {
		region = "ap-guangzhou"
	}

	signName := ""
	if req.Signature != nil {
		signName = req.Signature.SignatureCode
	}
	templateID := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		templateID = req.ChannelTemplateBinding.ProviderTemplate.TemplateCode
	}
	if templateID == "" {
		return &BatchSendResponse{Results: s.tencentFailAll(req.Tasks, "missing template_id", "CONFIG_ERROR")}, nil
	}

	// 逐号码手机号 + 共用模板参数（腾讯批量同模板同参数）
	phoneNumbers := make([]string, len(req.Tasks))
	for i, task := range req.Tasks {
		phoneNumbers[i] = task.Receiver
	}
	shared := req.MappedParams
	if len(req.TaskParams) > 0 && req.TaskParams[0] != nil {
		shared = req.TaskParams[0]
	}
	content := ""
	if req.ChannelTemplateBinding != nil && req.ChannelTemplateBinding.ProviderTemplate != nil {
		content = req.ChannelTemplateBinding.ProviderTemplate.TemplateContent
	}
	templateParamSet := tencentBuildParamsFromMap(shared, content)

	payload := map[string]any{
		"SmsSdkAppId":      sdkAppID,
		"SignName":         signName,
		"TemplateId":       templateID,
		"PhoneNumberSet":   phoneNumbers,
		"TemplateParamSet": templateParamSet,
	}
	payloadBytes, _ := json.Marshal(payload)

	timestamp := time.Now().Unix()
	date := time.Unix(timestamp, 0).UTC().Format("2006-01-02")
	authorization := tencentTC3Sign(secretID, secretKey, date, timestamp, payloadBytes)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, tencentSmsEndpoint, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return &BatchSendResponse{Results: s.tencentFailAll(req.Tasks, err.Error(), "HTTP_ERROR")}, nil
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")
	httpReq.Header.Set("Host", "sms.tencentcloudapi.com")
	httpReq.Header.Set("X-TC-Action", "SendSms")
	httpReq.Header.Set("X-TC-Version", tencentSmsVersion)
	httpReq.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	httpReq.Header.Set("X-TC-Region", region)
	httpReq.Header.Set("Authorization", authorization)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return &BatchSendResponse{Results: s.tencentFailAll(req.Tasks, err.Error(), "HTTP_ERROR")}, nil
	}
	defer resp.Body.Close()
	respBody := new(strings.Builder)
	_, _ = copyBuffer(respBody, resp.Body)
	responseData := jsonDump(map[string]any{"status": resp.StatusCode, "body": respBody.String()})

	var result struct {
		Response struct {
			SendStatusSet []struct {
				SerialNo string `json:"SerialNo"`
				Code     string `json:"Code"`
				Message  string `json:"Message"`
			} `json:"SendStatusSet"`
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	_ = json.Unmarshal([]byte(respBody.String()), &result)

	// 接口级错误：全部失败
	if result.Response.Error.Code != "" {
		return &BatchSendResponse{Results: s.tencentFailAll(req.Tasks, result.Response.Error.Message, result.Response.Error.Code)}, nil
	}

	// 逐号码结果映射回任务（SendStatusSet 与 PhoneNumberSet 顺序对应）
	results := make([]*SendResponse, len(req.Tasks))
	for i, task := range req.Tasks {
		if i < len(result.Response.SendStatusSet) {
			item := result.Response.SendStatusSet[i]
			if item.Code == "Ok" {
				results[i] = &SendResponse{
					Success: true, TaskID: task.TaskID, ProviderID: item.SerialNo,
					Status: string(model.PushTaskStatusSending), RequestData: string(payloadBytes), ResponseData: responseData,
				}
			} else {
				results[i] = &SendResponse{
					Success: false, TaskID: task.TaskID, ErrorCode: item.Code,
					ErrorMessage: item.Message, RequestData: string(payloadBytes), ResponseData: responseData,
				}
			}
		} else {
			results[i] = &SendResponse{
				Success: false, TaskID: task.TaskID, ErrorCode: "EMPTY_RESPONSE",
				ErrorMessage: "missing send status for task", RequestData: string(payloadBytes), ResponseData: responseData,
			}
		}
	}
	return &BatchSendResponse{Results: results}, nil
}

// tencentBuildParamsFromMap 从映射构造腾讯模板参数（按模板占位符顺序）。
// 腾讯 TemplateParamSet 为数组，顺序必须与模板占位符一致，否则内容错位；
// 无模板占位符时按 key 排序稳定回退。
func tencentBuildParamsFromMap(mapped map[string]string, content string) []string {
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

// tencentFailAll 为批量失败场景生成全部失败结果。
func (s *TencentSMSSender) tencentFailAll(tasks []*model.PushTask, errMsg, errCode string) []*SendResponse {
	results := make([]*SendResponse, len(tasks))
	for i, task := range tasks {
		results[i] = &SendResponse{Success: false, TaskID: task.TaskID, ErrorCode: errCode, ErrorMessage: errMsg}
	}
	return results
}
