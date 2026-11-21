package dto

import "time"

// ========== 回调日志查询 ==========

// CallbackListRequest 回调日志列表查询。
type CallbackListRequest struct {
	Page           int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize       int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	Type           string `form:"type" binding:"omitempty,oneof=report upstream"`
	TaskNo         string `form:"task_no"`
	RequestID      string `form:"request_id"`
	ProviderCode   string `form:"provider_code"`
	CallbackStatus string `form:"callback_status" binding:"omitempty,oneof=delivered failed rejected unknown"`
	StartDate      string `form:"start_date"` // YYYY-MM-DD
	EndDate        string `form:"end_date"`   // YYYY-MM-DD
}

// CallbackLogResponse 回调日志响应。
type CallbackLogResponse struct {
	ID                uint      `json:"id"`
	RequestID         string    `json:"request_id"`
	Type              string    `json:"type"`
	TaskNo            string    `json:"task_no"`
	AppID             uint      `json:"app_id"`
	ProviderCode      string    `json:"provider_code"`
	ProviderAccountID uint      `json:"provider_account_id"`
	ProviderID        string    `json:"provider_id"`
	Mobile            string    `json:"mobile"`
	Content           string    `json:"content"`
	CallbackStatus    string    `json:"callback_status"`
	ErrorCode         string    `json:"error_code"`
	ErrorMessage      string    `json:"error_message"`
	RawData           string    `json:"raw_data"`
	CreatedAt         time.Time `json:"created_at"`
}
