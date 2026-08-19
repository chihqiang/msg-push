package sender

import (
	"context"
	"time"
)

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

// DefaultCallbackResponse 各服务商约定的成功响应。
var (
	AliyunCallbackOK  = CallbackResponse{StatusCode: 200, Body: `{"code":0,"msg":"接收成功"}`}
	TencentCallbackOK = CallbackResponse{StatusCode: 200, Body: `{"result":0,"errmsg":"OK"}`}
	NeteaseCallbackOK = CallbackResponse{StatusCode: 200, Body: `{"code":200,"msg":"success"}`}
	GenericCallbackOK = CallbackResponse{StatusCode: 200, Body: `{"code":0,"message":"ok"}`}
)
