// Package sender 提供消息发送器层：定义 Sender/StatusQuerier/BatchSender 接口与 8 个服务商实现（供 worker 消费端调用）。
package sender

import (
	"context"
	"time"

	"chihqiang/msg-push/model"
)

// SendRequest 发送请求。
type SendRequest struct {
	Task                   *model.PushTask
	ProviderAccount        *model.ProviderAccount
	ChannelTemplateBinding *model.ChannelTemplateBinding
	Signature              *model.ProviderSignature
	MappedParams           map[string]string // 供应商变量名到值的映射
	RenderedContent        string            // 供应商模板渲染后的内容

	// 手机号预解析字段（仅 SMS 类型，worker 填充）
	PhoneRegion         string
	PhoneCountryCode    string
	PhoneNationalNumber string
	PhoneE164           string
}

// smsReceiver 选择发送用手机号（短信服务商通用）。
// 规则：国内号（region=CN）用 11 位 national；国际号用 E.164（+<国家码><号码>）；
// 均未解析出时回退到原始 receiver。
func smsReceiver(req *SendRequest) string {
	if req == nil {
		return ""
	}
	if req.PhoneRegion == "CN" && req.PhoneNationalNumber != "" {
		return req.PhoneNationalNumber
	}
	if req.PhoneE164 != "" {
		return req.PhoneE164
	}
	if req.Task != nil {
		return req.Task.Receiver
	}
	return ""
}

// SendResponse 发送响应。
type SendResponse struct {
	Success      bool
	ProviderID   string // 服务商消息ID
	ErrorCode    string
	ErrorMessage string
	TaskID       string
	Status       string // "sending"(等待回执) 或 "success"(直接完成)
	RequestData  string // 请求参数(JSON)
	ResponseData string // 响应数据(JSON)
}

// Sender 发送器接口。
type Sender interface {
	Send(ctx context.Context, req *SendRequest) (*SendResponse, error)
	GetProviderCode() string
}

// Resolver 发送器解析器。
type Resolver interface {
	GetSender(providerCode string) (Sender, error)
	GetBatchSender(providerCode string) (BatchSender, error)
}

// StatusQueryRequest 状态查询请求（主动回执查询）。
type StatusQueryRequest struct {
	Task            *model.PushTask
	ProviderAccount *model.ProviderAccount
	ProviderMsgID   string    // 服务商消息ID（发送响应中的 ProviderID）
	PhoneNumber     string    // 接收手机号
	SendDate        time.Time // 发送日期（按天查询）
}

// StatusQueryResult 单条状态查询结果。
type StatusQueryResult struct {
	ProviderMsgID string
	PhoneNumber   string
	Status        string // delivered / failed / unknown
	ErrorCode     string
	ErrorMessage  string
}

// StatusQueryResponse 状态查询响应。
type StatusQueryResponse struct {
	Results []*StatusQueryResult
}

// StatusQuerier 状态查询器接口（短信类服务商可选实现，用于主动补单回执）。
type StatusQuerier interface {
	QueryStatus(ctx context.Context, req *StatusQueryRequest) (*StatusQueryResponse, error)
	GetProviderCode() string
}

// BatchSendRequest 批量发送请求（批量发送）。
// 所有任务共用一个通道绑定/签名/服务商账号；TaskParams 与 Tasks 一一对应，
// 存放每个任务解析后的供应商变量映射（为空项回退到 MappedParams）。
type BatchSendRequest struct {
	Tasks                  []*model.PushTask
	ProviderAccount        *model.ProviderAccount
	ChannelTemplateBinding *model.ChannelTemplateBinding
	Signature              *model.ProviderSignature
	MappedParams           map[string]string   // 共用映射参数（无独立参数时使用）
	TaskParams             []map[string]string // 每任务独立映射参数（可空，与 Tasks 对齐）
	RenderedContent        string              // 供应商模板渲染后的内容
}

// BatchSendResponse 批量发送响应，Results 与 Tasks 一一对应。
type BatchSendResponse struct {
	Results []*SendResponse
}

// BatchSender 批量发送器接口（可选实现，用于批量消息聚合调用服务商批量 API）。
type BatchSender interface {
	Sender
	// BatchSend 批量发送消息（一次 API 调用发送多个号码）。
	BatchSend(ctx context.Context, req *BatchSendRequest) (*BatchSendResponse, error)
	// SupportsBatchSend 是否支持批量发送。
	SupportsBatchSend() bool
}
