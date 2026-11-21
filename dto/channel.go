package dto

import "time"

// ChannelType 通道类型枚举，与 model.ChannelType 对齐。
const (
	ChannelTypeSMS      = "sms"
	ChannelTypeEmail    = "email"
	ChannelTypeWeCom    = "wecom"
	ChannelTypeDingTalk = "dingtalk"
)

// ChannelListRequest 通道列表查询。
type ChannelListRequest struct {
	PageRequest
	Type string `form:"type" binding:"omitempty,oneof=sms email wecom dingtalk"` // 通道类型过滤
	Key  string `form:"key" binding:"omitempty,max=100"`                         // 关键字：编码/名称
}

// CreateChannelRequest 创建通道请求。
type CreateChannelRequest struct {
	Code   string `json:"code" binding:"required,max=64"`
	Name   string `json:"name" binding:"required,max=64"`
	Type   string `json:"type" binding:"required,oneof=sms email wecom dingtalk"`
	Config string `json:"config"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateChannelRequest 更新通道请求。Code 一旦分配不可修改。
type UpdateChannelRequest struct {
	Name   string `json:"name" binding:"omitempty,max=64"`
	Type   string `json:"type" binding:"omitempty,oneof=sms email wecom dingtalk"`
	Config string `json:"config"`
	Status *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
}

// ChannelResponse 通道响应。
type ChannelResponse struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Config    string    `json:"config,omitempty"`
	Status    int8      `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}
