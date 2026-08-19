package dto

import "time"

// ========== 日志查询 ==========

// LogListRequest 推送日志列表查询。
type LogListRequest struct {
	Page              int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize          int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	TaskNo            string `form:"task_no"`
	RequestID         string `form:"request_id"`
	AppID             uint   `form:"app_id"`
	ProviderAccountID uint   `form:"provider_account_id"`
	Status            string `form:"status" binding:"omitempty,oneof=pending sending success failed"`
	StartDate         string `form:"start_date"` // YYYY-MM-DD
	EndDate           string `form:"end_date"`   // YYYY-MM-DD
}

// LogResponse 推送日志响应。
type LogResponse struct {
	ID                uint      `json:"id"`
	RequestID         string    `json:"request_id"`
	TaskID            uint      `json:"task_id"`
	TaskNo            string    `json:"task_no"`
	AppID             uint      `json:"app_id"`
	ProviderAccountID uint      `json:"provider_account_id"`
	ProviderMsgID     string    `json:"provider_msg_id"`
	Receiver          string    `json:"receiver"`
	Status            string    `json:"status"`
	ProviderResp      string    `json:"provider_resp"`
	ErrorCode         string    `json:"error_code"`
	ErrorMsg          string    `json:"error_msg"`
	CostTime          int       `json:"cost_time"`
	CreatedAt         time.Time `json:"created_at"`
}
