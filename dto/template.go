package dto

import "time"

// TemplateListRequest 模板列表查询。
type TemplateListRequest struct {
	PageRequest
	ChannelID uint   `form:"channel_id"`                      // 通道过滤
	Key       string `form:"key" binding:"omitempty,max=100"` // 关键字：编码/名称
}

// CreateTemplateRequest 创建模板请求。
type CreateTemplateRequest struct {
	Code      string `json:"code" binding:"required,max=64"`
	ChannelID uint   `json:"channel_id" binding:"required"`
	Name      string `json:"name" binding:"required,max=128"`
	Content   string `json:"content" binding:"required"`
	Signature string `json:"signature" binding:"omitempty,max=64"`
	Remark    string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateTemplateRequest 更新模板请求。Code 一旦分配不可修改。
type UpdateTemplateRequest struct {
	ChannelID uint   `json:"channel_id"`
	Name      string `json:"name" binding:"omitempty,max=128"`
	Content   string `json:"content"`
	Signature string `json:"signature" binding:"omitempty,max=64"`
	Status    *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Remark    string `json:"remark" binding:"omitempty,max=255"`
}

// TemplateResponse 模板响应。
type TemplateResponse struct {
	ID        uint      `json:"id"`
	Code      string    `json:"code"`
	ChannelID uint      `json:"channel_id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Signature string    `json:"signature"`
	Status    int8      `json:"status"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}
