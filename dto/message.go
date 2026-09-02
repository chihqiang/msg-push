// Package dto 定义 HTTP 接口层的请求与响应数据结构。
package dto

import "time"

// PageRequest 分页请求。
type PageRequest struct {
	Page     int `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}

// SendRequest 单条发送请求。
// ChannelCode 与 TemplateCode 至少传一个：只传 TemplateCode 时自动定位模板所属通道；
// 两者都传时校验模板属于该通道；只传 ChannelCode 表示不套模板直接发送。
type SendRequest struct {
	ChannelCode    string            `json:"channel_code" binding:"omitempty"`
	TemplateCode   string            `json:"template_code"`
	Receiver       string            `json:"receiver" binding:"required"`
	TemplateParams map[string]string `json:"template_params"`
	SignatureName  string            `json:"signature_name" binding:"omitempty,max=200"`
	ScheduledAt    *time.Time        `json:"scheduled_at"`
}

// BatchSendRequest 批量发送请求。
// ChannelCode 与 TemplateCode 至少传一个，规则同单条发送。
// Receivers 上限 1000：防止单条请求创建海量任务、绕过每日配额（配额按接收者数扣减）。
type BatchSendRequest struct {
	ChannelCode    string            `json:"channel_code" binding:"omitempty"`
	TemplateCode   string            `json:"template_code"`
	Receivers      []string          `json:"receivers" binding:"required,min=1,max=1000"`
	TemplateParams map[string]string `json:"template_params"`
	SignatureName  string            `json:"signature_name" binding:"omitempty,max=200"`
	ScheduledAt    *time.Time        `json:"scheduled_at"`
}

// SendResponse 单条发送响应。
type SendResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	IsTest    bool      `json:"is_test"`
	CreatedAt time.Time `json:"created_at"`
}

// BatchSendResponse 批量发送响应。
type BatchSendResponse struct {
	BatchID      string `json:"batch_id"`
	TotalCount   int    `json:"total_count"`
	SuccessCount int    `json:"success_count"`
	FailedCount  int    `json:"failed_count"`
	IsTest       bool   `json:"is_test"`
}
