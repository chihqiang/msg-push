package dto

import (
	"time"

	"chihqiang/msg-push/model"
)

// ========== 通道-模板绑定 ==========

// ChannelBindingListRequest 绑定列表查询。
type ChannelBindingListRequest struct {
	Page     int `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}

// CreateChannelBindingRequest 创建绑定。
type CreateChannelBindingRequest struct {
	ProviderTemplateID   uint                     `json:"provider_template_id" binding:"required"`
	ProviderID           uint                     `json:"provider_id" binding:"required"`
	ParamMapping         []model.ParamMappingItem `json:"param_mapping"`
	Weight               *int                     `json:"weight" binding:"omitempty,min=1,max=100"`
	Priority             *int                     `json:"priority" binding:"omitempty,min=0,max=1000"`
	Status               *int8                    `json:"status" binding:"omitempty,oneof=0 1"`
	IsActive             *int8                    `json:"is_active" binding:"omitempty,oneof=0 1"`
	AutoDisableOnFail    *bool                    `json:"auto_disable_on_fail"`
	AutoDisableThreshold *int                     `json:"auto_disable_threshold" binding:"omitempty,min=1,max=100"`
}

// UpdateChannelBindingRequest 更新绑定。
type UpdateChannelBindingRequest struct {
	ProviderTemplateID   *uint                    `json:"provider_template_id"`
	ProviderID           *uint                    `json:"provider_id"`
	ParamMapping         []model.ParamMappingItem `json:"param_mapping"`
	Weight               *int                     `json:"weight" binding:"omitempty,min=1,max=100"`
	Priority             *int                     `json:"priority" binding:"omitempty,min=0,max=1000"`
	Status               *int8                    `json:"status" binding:"omitempty,oneof=0 1"`
	IsActive             *int8                    `json:"is_active" binding:"omitempty,oneof=0 1"`
	AutoDisableOnFail    *bool                    `json:"auto_disable_on_fail"`
	AutoDisableThreshold *int                     `json:"auto_disable_threshold" binding:"omitempty,min=1,max=100"`
}

// ChannelBindingResponse 绑定响应。
type ChannelBindingResponse struct {
	ID                   uint                     `json:"id"`
	ChannelID            uint                     `json:"channel_id"`
	ProviderTemplateID   uint                     `json:"provider_template_id"`
	ProviderTemplateName string                   `json:"provider_template_name"`
	ProviderID           uint                     `json:"provider_id"`
	ProviderName         string                   `json:"provider_name"`
	ProviderType         string                   `json:"provider_type"`
	ParamMapping         []model.ParamMappingItem `json:"param_mapping"`
	Weight               int                      `json:"weight"`
	Priority             int                      `json:"priority"`
	Status               int8                     `json:"status"`
	IsActive             int8                     `json:"is_active"`
	AutoDisableOnFail    bool                     `json:"auto_disable_on_fail"`
	AutoDisableThreshold int                      `json:"auto_disable_threshold"`
	CreatedAt            time.Time                `json:"created_at"`
}

// AvailableProviderTemplateResponse 可用供应商模板（用于绑定下拉）。
type AvailableProviderTemplateResponse struct {
	ID              uint     `json:"id"`
	TemplateCode    string   `json:"template_code"`
	TemplateName    string   `json:"template_name"`
	TemplateContent string   `json:"template_content"`
	Variables       []string `json:"variables"`
	ProviderID      uint     `json:"provider_id"`
	ProviderCode    string   `json:"provider_code"`
	ProviderType    string   `json:"provider_type"`
	Status          int8     `json:"status"`
}
