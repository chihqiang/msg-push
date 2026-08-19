package dto

import "time"

// AvailableProvidersRequest 可用服务商列表查询。
type AvailableProvidersRequest struct {
	ProviderType string `form:"provider_type" binding:"omitempty,oneof=sms email wechat_work dingtalk"` // 消息类型过滤
}

// ========== 服务商账号 ==========

// ProviderAccountListRequest 服务商账号列表查询。
type ProviderAccountListRequest struct {
	Page         int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize     int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	ProviderType string `form:"provider_type" binding:"omitempty,oneof=sms email wechat_work dingtalk"`
	Status       *int8  `form:"status" binding:"omitempty,oneof=0 1"`
	Key          string `form:"key" binding:"omitempty,max=100"` // 关键字：账号名称/账号代码/服务商代码
}

// CreateProviderAccountRequest 创建服务商账号。
type CreateProviderAccountRequest struct {
	AccountName  string         `json:"account_name" binding:"required,min=2,max=100"`
	ProviderCode string         `json:"provider_code" binding:"required"`
	Config       map[string]any `json:"config" binding:"required"`
	Status       *int8          `json:"status" binding:"omitempty,oneof=0 1"`
	Remark       string         `json:"remark" binding:"omitempty,max=255"`
}

// UpdateProviderAccountRequest 更新服务商账号。
type UpdateProviderAccountRequest struct {
	AccountName string         `json:"account_name" binding:"omitempty,min=2,max=100"`
	Config      map[string]any `json:"config"`
	Status      *int8          `json:"status" binding:"omitempty,oneof=0 1"`
	Remark      string         `json:"remark" binding:"omitempty,max=255"`
}

// ProviderAccountResponse 服务商账号响应。
type ProviderAccountResponse struct {
	ID           uint           `json:"id"`
	AccountCode  string         `json:"account_code"`
	AccountName  string         `json:"account_name"`
	ProviderCode string         `json:"provider_code"`
	ProviderName string         `json:"provider_name"`
	ProviderType string         `json:"provider_type"`
	Config       map[string]any `json:"config"`
	Status       int8           `json:"status"`
	Remark       string         `json:"remark"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// ProviderAccountListResponse 服务商账号列表响应。
type ProviderAccountListResponse struct {
	List     []*ProviderAccountResponse `json:"list"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}

// AvailableProviderResponse 可用服务商响应。
type AvailableProviderResponse struct {
	Code              string   `json:"code"`
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Description       string   `json:"description"`
	SupportsSend      bool     `json:"supports_send"`
	SupportsBatchSend bool     `json:"supports_batch_send"`
	SupportsCallback  bool     `json:"supports_callback"`
	RequiresSignature bool     `json:"requires_signature"`
	Website           string   `json:"website"`
	ConsoleUrl        string   `json:"console_url"`
	SortOrder         int      `json:"sort_order"`
	Tags              []string `json:"tags"`
	// ConfigFields 配置字段定义（前端据此动态渲染"新建服务商账号"表单）。
	ConfigFields []*ConfigFieldResponse `json:"config_fields"`
}

// ConfigFieldResponse 配置字段响应。
type ConfigFieldResponse struct {
	Key            string `json:"key"`
	Label          string `json:"label"`
	Description    string `json:"description"`
	Type           string `json:"type"`
	Required       bool   `json:"required"`
	Example        string `json:"example"`
	Placeholder    string `json:"placeholder"`
	ValidationRule string `json:"validation_rule"`
	DefaultValue   string `json:"default_value"`
	Options        []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `json:"options"`
}

// TestProviderRequest 测试服务商请求。
type TestProviderRequest struct {
	AccountID uint   `json:"account_id" binding:"required"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Message   string `json:"message"`
}
