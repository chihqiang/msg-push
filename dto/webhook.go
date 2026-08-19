package dto

import "time"

// ========== Webhook 配置 ==========

// WebhookConfigListRequest Webhook 配置列表查询。
type WebhookConfigListRequest struct {
	Page     int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
	Key      string `form:"key" binding:"omitempty,max=100"` // 关键字：名称/回调地址
}

// CreateWebhookConfigRequest 创建 Webhook 配置。
type CreateWebhookConfigRequest struct {
	Name        string `json:"name" binding:"required,max=100"`
	AppID       uint   `json:"app_id"`
	WebhookURL  string `json:"webhook_url" binding:"required,url,max=500"`
	Secret      string `json:"secret" binding:"omitempty,max=128"`
	Events      string `json:"events" binding:"omitempty,max=200"`
	Status      *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	RetryCount  *int   `json:"retry_count" binding:"omitempty,min=0,max=10"`
	Timeout     *int   `json:"timeout" binding:"omitempty,min=1,max=60"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// UpdateWebhookConfigRequest 更新 Webhook 配置。
type UpdateWebhookConfigRequest struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	AppID       *uint  `json:"app_id"`
	WebhookURL  string `json:"webhook_url" binding:"omitempty,url,max=500"`
	Secret      string `json:"secret" binding:"omitempty,max=128"`
	Events      string `json:"events" binding:"omitempty,max=200"`
	Status      *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	RetryCount  *int   `json:"retry_count" binding:"omitempty,min=0,max=10"`
	Timeout     *int   `json:"timeout" binding:"omitempty,min=1,max=60"`
	Description string `json:"description" binding:"omitempty,max=255"`
}

// WebhookConfigResponse Webhook 配置响应。
type WebhookConfigResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	AppID       uint      `json:"app_id"`
	WebhookURL  string    `json:"webhook_url"`
	Events      string    `json:"events"`
	Status      int8      `json:"status"`
	RetryCount  int       `json:"retry_count"`
	Timeout     int       `json:"timeout"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ========== Webhook 日志查询 ==========

// WebhookLogListRequest Webhook 日志列表查询。
type WebhookLogListRequest struct {
	Page      int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	TaskNo    string `form:"task_no"`
	RequestID string `form:"request_id"`
	Event     string `form:"event"`
	Status    string `form:"status" binding:"omitempty,oneof=pending success failed"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// WebhookLogResponse Webhook 日志响应。
type WebhookLogResponse struct {
	ID              uint      `json:"id"`
	RequestID       string    `json:"request_id"`
	TaskNo          string    `json:"task_no"`
	AppID           uint      `json:"app_id"`
	WebhookConfigID uint      `json:"webhook_config_id"`
	WebhookURL      string    `json:"webhook_url"`
	Event           string    `json:"event"`
	RequestData     string    `json:"request_data"`
	ResponseStatus  int       `json:"response_status"`
	ResponseData    string    `json:"response_data"`
	Status          string    `json:"status"`
	ErrorMessage    string    `json:"error_message"`
	RetryCount      int       `json:"retry_count"`
	MaxRetries      int       `json:"max_retries"`
	TimeoutSeconds  int       `json:"timeout_seconds"`
	CreatedAt       time.Time `json:"created_at"`
}
