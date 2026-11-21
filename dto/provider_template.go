package dto

import "time"

// ========== 供应商模板 ==========

// ProviderTemplateListRequest 供应商模板列表查询。
type ProviderTemplateListRequest struct {
	ProviderID *uint  `form:"provider_id"`
	Status     *int8  `form:"status" binding:"omitempty,oneof=0 1"`
	Key        string `form:"key" binding:"omitempty,max=100"` // 关键字：模板名称/模板编码
	Page       int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}

// CreateProviderTemplateRequest 创建供应商模板。
type CreateProviderTemplateRequest struct {
	ProviderID      uint     `json:"provider_id" binding:"required"`
	TemplateCode    string   `json:"template_code" binding:"required,max=100"`
	TemplateName    string   `json:"template_name" binding:"required,max=200"`
	ContentType     string   `json:"content_type" binding:"omitempty,oneof=text html markdown"`
	TemplateContent string   `json:"template_content"`
	Variables       []string `json:"variables"`
	Status          *int8    `json:"status" binding:"omitempty,oneof=0 1"`
	Remark          string   `json:"remark" binding:"omitempty,max=255"`
}

// UpdateProviderTemplateRequest 更新供应商模板。
type UpdateProviderTemplateRequest struct {
	TemplateCode    string   `json:"template_code" binding:"omitempty,max=100"`
	TemplateName    string   `json:"template_name" binding:"omitempty,max=200"`
	ContentType     string   `json:"content_type" binding:"omitempty,oneof=text html markdown"`
	TemplateContent string   `json:"template_content"`
	Variables       []string `json:"variables"`
	Status          *int8    `json:"status" binding:"omitempty,oneof=0 1"`
	Remark          string   `json:"remark" binding:"omitempty,max=255"`
}

// ProviderTemplateResponse 供应商模板响应。
type ProviderTemplateResponse struct {
	ID              uint      `json:"id"`
	ProviderID      uint      `json:"provider_id"`
	TemplateCode    string    `json:"template_code"`
	TemplateName    string    `json:"template_name"`
	ContentType     string    `json:"content_type"`
	TemplateContent string    `json:"template_content"`
	Variables       []string  `json:"variables"`
	Status          int8      `json:"status"`
	Remark          string    `json:"remark"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
