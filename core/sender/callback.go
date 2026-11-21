package sender

import (
	"context"
	"strings"
	"time"

	"chihqiang/msg-push/model"
)

// ==================== 回调接口与结构体 ====================

// CallbackRequest 服务商回调请求。
type CallbackRequest struct {
	ProviderCode      string
	ProviderAccountID uint
	RawBody           []byte
	Headers           map[string]string
	QueryParams       map[string]string
	FormData          map[string]string
}

// CallbackResult 服务商回调解析结果。
type CallbackResult struct {
	Type         string // report / upstream
	ProviderID   string // 服务商消息ID（回执关联用）
	Status       string // delivered / failed / rejected
	ErrorCode    string
	ErrorMessage string
	Mobile       string
	Content      string
	ReportTime   time.Time
}

// CallbackResponse 返回给服务商的响应。
type CallbackResponse struct {
	StatusCode int
	Body       string
}

// CallbackParser 服务商定制回调解析器接口。
// 各服务商回调格式不同（阿里云数组、腾讯云数组、网易 eventType 对象），由实现按格式解析。
type CallbackParser interface {
	// Parse 解析服务商回调 body，返回一个或多个回调结果。
	Parse(ctx context.Context, req *CallbackRequest) (CallbackResponse, []*CallbackResult, error)
	// GetProviderCode 服务商代码。
	GetProviderCode() string
}

// 各服务商约定的成功响应。
var (
	AliyunCallbackOK  = CallbackResponse{StatusCode: 200, Body: `{"code":0,"msg":"接收成功"}`}
	TencentCallbackOK = CallbackResponse{StatusCode: 200, Body: `{"result":0,"errmsg":"OK"}`}
	NeteaseCallbackOK = CallbackResponse{StatusCode: 200, Body: `{"code":200,"msg":"success"}`}
	GenericCallbackOK = CallbackResponse{StatusCode: 200, Body: `{"code":0,"message":"ok"}`}
)

// normalizeCallbackStatus 通用状态规范化（供解析器复用）。
// 注意：先判失败/拒绝（如 "undelivered" 含子串 "deliver" 但语义为失败），
// 再判送达（"delivered"/"DELIVRD" 等，取子串 "deliv" 覆盖送达码缺失字母 e 的变体）。
func normalizeCallbackStatus(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case s == "failed" || s == "fail" || s == "undelivered" || s == "undeliverd" || s == "undelivrd":
		return model.CallbackStatusFailed
	case s == "rejected" || s == "refuse" || s == "reject":
		return model.CallbackStatusRejected
	case s == "success" || s == "sent" || strings.Contains(s, "deliv"):
		return model.CallbackStatusDelivered
	default:
		return model.CallbackStatusUnknown
	}
}

// ==================== 回调解析器注册表 ====================

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

// init 注册内置回调解析器（各解析器实现见对应服务商文件）。
func init() {
	RegisterCallbackParser(&AliyunCallbackParser{})
	RegisterCallbackParser(&TencentCallbackParser{})
	RegisterCallbackParser(&NeteaseCallbackParser{})
}
