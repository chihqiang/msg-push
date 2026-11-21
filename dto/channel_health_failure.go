package dto

import (
	"time"

	"chihqiang/msg-push/model"
)

// ========== 通道测试发送 + 健康历史 ==========

// TestChannelRequest 通道测试发送请求。
type TestChannelRequest struct {
	Receiver string `json:"receiver" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

// TestChannelResponse 通道测试发送响应。
type TestChannelResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

// ChannelHealthHistoryResponse 通道健康历史响应。
type ChannelHealthHistoryResponse struct {
	ID                uint                      `json:"id"`
	ProviderChannelID uint                      `json:"provider_channel_id"`
	CheckTime         time.Time                 `json:"check_time"`
	Status            model.ChannelHealthStatus `json:"status"`
	ResponseTime      int                       `json:"response_time"`
	ErrorCount        int                       `json:"error_count"`
	SuccessRate       float64                   `json:"success_rate"`
	IsAvailable       int8                      `json:"is_available"`
}

// ========== 失败规则 ==========

// FailureRuleListRequest 失败规则列表查询。
type FailureRuleListRequest struct {
	Page     int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
	Scene    string `form:"scene" binding:"omitempty,oneof=send_failure callback_failure"`
	Status   *int8  `form:"status" binding:"omitempty,oneof=0 1"`
	Key      string `form:"key" binding:"omitempty,max=100"` // 关键字：规则名称
}

// CreateFailureRuleRequest 创建失败规则。
type CreateFailureRuleRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=100"`
	Scene        string `json:"scene" binding:"required,oneof=send_failure callback_failure"`
	ProviderCode string `json:"provider_code" binding:"omitempty,max=50"`
	MessageType  string `json:"message_type" binding:"omitempty,oneof=sms email wechat_work dingtalk"`
	ErrorCode    string `json:"error_code" binding:"omitempty,max=200"`
	ErrorKeyword string `json:"error_keyword" binding:"omitempty,max=200"`
	Action       string `json:"action" binding:"required,oneof=retry switch_provider fail alert"`
	ActionConfig string `json:"action_config"`
	Priority     int    `json:"priority" binding:"omitempty,min=0,max=1000"`
	Status       *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Remark       string `json:"remark" binding:"omitempty,max=500"`
}

// UpdateFailureRuleRequest 更新失败规则。
type UpdateFailureRuleRequest struct {
	Name         string `json:"name" binding:"omitempty,min=2,max=100"`
	Scene        string `json:"scene" binding:"omitempty,oneof=send_failure callback_failure"`
	ProviderCode string `json:"provider_code" binding:"omitempty,max=50"`
	MessageType  string `json:"message_type" binding:"omitempty,oneof=sms email wechat_work dingtalk"`
	ErrorCode    string `json:"error_code" binding:"omitempty,max=200"`
	ErrorKeyword string `json:"error_keyword" binding:"omitempty,max=200"`
	Action       string `json:"action" binding:"omitempty,oneof=retry switch_provider fail alert"`
	ActionConfig string `json:"action_config"`
	Priority     *int   `json:"priority" binding:"omitempty,min=0,max=1000"`
	Status       *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Remark       string `json:"remark" binding:"omitempty,max=500"`
}

// FailureRuleResponse 失败规则响应。
type FailureRuleResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Scene        string    `json:"scene"`
	ProviderCode string    `json:"provider_code"`
	MessageType  string    `json:"message_type"`
	ErrorCode    string    `json:"error_code"`
	ErrorKeyword string    `json:"error_keyword"`
	Action       string    `json:"action"`
	ActionConfig string    `json:"action_config"`
	Priority     int       `json:"priority"`
	Status       int8      `json:"status"`
	Remark       string    `json:"remark"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OptionItem 选项项（用于规则选项）。
type OptionItem struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FailureRuleOptionsResponse 失败规则选项响应。
type FailureRuleOptionsResponse struct {
	Scenes  []OptionItem `json:"scenes"`
	Actions []OptionItem `json:"actions"`
}
