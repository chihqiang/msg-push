package dto

import "time"

// ========== 任务查询 ==========

// PushTaskListRequest 任务列表查询。
type PushTaskListRequest struct {
	Page      int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	TaskNo    string `form:"task_no"`
	RequestID string `form:"request_id"`
	BatchID   string `form:"batch_id"`
	AppID     uint   `form:"app_id"`
	ChannelID uint   `form:"channel_id"`
	Receiver  string `form:"receiver"`
	Status    string `form:"status" binding:"omitempty,oneof=pending sending success failed"`
	StartDate string `form:"start_date"` // YYYY-MM-DD
	EndDate   string `form:"end_date"`   // YYYY-MM-DD
}

// PushTaskResponse 任务响应。
type PushTaskResponse struct {
	ID          uint       `json:"id"`
	TaskID      string     `json:"task_id"`
	RequestID   string     `json:"request_id"`
	AppID       uint       `json:"app_id"`
	BatchID     string     `json:"batch_id"`
	ChannelID   uint       `json:"channel_id"`
	TemplateID  uint       `json:"template_id"`
	Receiver    string     `json:"receiver"`
	Params      string     `json:"params"`
	Signature   string     `json:"signature"`
	IsTest      bool       `json:"is_test"`
	Status      string     `json:"status"`
	ErrorMsg    string     `json:"error_msg"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// 关联数据
	ChannelName string `json:"channel_name,omitempty"`
}

// ========== 批量任务查询 ==========

// PushBatchTaskListRequest 批次列表查询。
type PushBatchTaskListRequest struct {
	Page      int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	BatchID   string `form:"batch_id"`
	AppID     uint   `form:"app_id"`
	Status    string `form:"status" binding:"omitempty,oneof=processing completed failed"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// PushBatchTaskResponse 批次响应。
type PushBatchTaskResponse struct {
	ID           uint      `json:"id"`
	AppID        uint      `json:"app_id"`
	BatchID      string    `json:"batch_id"`
	ChannelID    uint      `json:"channel_id"`
	TemplateID   uint      `json:"template_id"`
	TotalCount   int       `json:"total_count"`
	SuccessCount int       `json:"success_count"`
	FailedCount  int       `json:"failed_count"`
	PendingCount int       `json:"pending_count"`
	IsTest       bool      `json:"is_test"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	// 计算字段
	CompletionRate float64 `json:"completion_rate"`
}
