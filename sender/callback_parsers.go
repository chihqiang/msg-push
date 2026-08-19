package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// ==================== 阿里云短信 ====================

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

// ==================== 腾讯云短信 ====================

// TencentCallbackParser 腾讯云短信回执解析器。
// 格式：JSON 数组 [{"user_receive_time":"...","nationcode":"86","mobile":"138****","report_status":"SUCCESS","errmsg":"DELIVRD","description":"...","sid":"xxx"}]
type TencentCallbackParser struct{}

// GetProviderCode 返回服务商代码。
func (p *TencentCallbackParser) GetProviderCode() string {
	return CodeTencentSMS
}

// Parse 解析腾讯云回执。
func (p *TencentCallbackParser) Parse(ctx context.Context, req *CallbackRequest) (CallbackResponse, []*CallbackResult, error) {
	var reports []struct {
		UserReceiveTime string `json:"user_receive_time"`
		NationCode      string `json:"nationcode"`
		Mobile          string `json:"mobile"`
		ReportStatus    string `json:"report_status"`
		ErrMsg          string `json:"errmsg"`
		Description     string `json:"description"`
		SID             string `json:"sid"`
	}
	if err := json.Unmarshal(req.RawBody, &reports); err != nil {
		return TencentCallbackOK, nil, fmt.Errorf("invalid tencent callback: %w", err)
	}

	results := make([]*CallbackResult, 0, len(reports))
	for _, r := range reports {
		status := model.CallbackStatusDelivered
		if r.ReportStatus != "SUCCESS" {
			status = model.CallbackStatusFailed
		}
		reportTime, _ := time.ParseInLocation("2006-01-02 15:04:05", r.UserReceiveTime, time.Local)
		results = append(results, &CallbackResult{
			Type:         string(model.CallbackTypeReport),
			ProviderID:   r.SID,
			Status:       status,
			ErrorCode:    r.ErrMsg,
			ErrorMessage: r.Description,
			Mobile:       r.Mobile,
			ReportTime:   reportTime,
		})
	}
	return TencentCallbackOK, results, nil
}

// ==================== 网易云信短信 ====================

// NeteaseCallbackParser 网易云信短信回执解析器。
// 下行回执（eventType=11）：{"eventType":"11","objects":[{"mobile":"...","sendid":"1490","result":"DELIVRD","reason":"...","reportTime":"..."}]}
// 上行短信（eventType=12）：{"eventType":"12","objects":[{"mobile":"...","content":"回复内容","receiveTime":"..."}]}
type NeteaseCallbackParser struct{}

// GetProviderCode 返回服务商代码。
func (p *NeteaseCallbackParser) GetProviderCode() string {
	return CodeNeteaseSMS
}

// Parse 解析网易回执。
func (p *NeteaseCallbackParser) Parse(ctx context.Context, req *CallbackRequest) (CallbackResponse, []*CallbackResult, error) {
	var payload struct {
		EventType string `json:"eventType"`
		Objects   []struct {
			Mobile      json.RawMessage `json:"mobile"`
			SendID      json.RawMessage `json:"sendid"`
			Result      string          `json:"result"`
			ReportTime  string          `json:"reportTime"`
			Reason      string          `json:"reason"`
			Content     string          `json:"content"`
			ReceiveTime string          `json:"receiveTime"`
		} `json:"objects"`
	}
	if err := json.Unmarshal(req.RawBody, &payload); err != nil {
		return NeteaseCallbackOK, nil, fmt.Errorf("invalid netease callback: %w", err)
	}

	switch payload.EventType {
	case "11":
		// 下行投递回执
		results := make([]*CallbackResult, 0, len(payload.Objects))
		for _, obj := range payload.Objects {
			sendID := neteaseFlexString(rawToAny(obj.SendID))
			if sendID == "" {
				continue
			}
			status := model.CallbackStatusDelivered
			if obj.Result != "DELIVRD" {
				status = model.CallbackStatusFailed
			}
			reportTime, _ := time.ParseInLocation("2006-01-02 15:04:05", obj.ReportTime, time.Local)
			results = append(results, &CallbackResult{
				Type:         string(model.CallbackTypeReport),
				ProviderID:   sendID,
				Status:       status,
				ErrorMessage: obj.Reason,
				Mobile:       neteaseFlexString(rawToAny(obj.Mobile)),
				ReportTime:   reportTime,
			})
		}
		return NeteaseCallbackOK, results, nil
	case "12":
		// 上行短信
		results := make([]*CallbackResult, 0, len(payload.Objects))
		for _, obj := range payload.Objects {
			results = append(results, &CallbackResult{
				Type:    string(model.CallbackTypeUpstream),
				Status:  model.CallbackStatusDelivered,
				Mobile:  neteaseFlexString(rawToAny(obj.Mobile)),
				Content: obj.Content,
			})
		}
		return NeteaseCallbackOK, results, nil
	default:
		return NeteaseCallbackOK, nil, nil
	}
}

// ==================== 解析器注册 ====================

// 全局回调解析器注册表。
var callbackParsers = map[string]CallbackParser{}

// RegisterCallbackParser 注册解析器。
func RegisterCallbackParser(p CallbackParser) {
	callbackParsers[p.GetProviderCode()] = p
}

// GetCallbackParser 获取服务商回调解析器（未注册返回 nil）。
func GetCallbackParser(providerCode string) CallbackParser {
	return callbackParsers[providerCode]
}

// init 注册内置解析器。
func init() {
	RegisterCallbackParser(&AliyunCallbackParser{})
	RegisterCallbackParser(&TencentCallbackParser{})
	RegisterCallbackParser(&NeteaseCallbackParser{})
}

// normalizeCallbackStatus 通用状态规范化（供解析器复用）。
func normalizeCallbackStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.Contains(s, "deliver") || s == "success" || s == "sent":
		return model.CallbackStatusDelivered
	case s == "failed" || s == "fail" || s == "undelivered":
		return model.CallbackStatusFailed
	case s == "rejected" || s == "refuse":
		return model.CallbackStatusRejected
	default:
		return model.CallbackStatusUnknown
	}
}
