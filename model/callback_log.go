package model

import "time"

// CallbackType 回调类型。
type CallbackType string

// 回调类型常量。
const (
	CallbackTypeReport   CallbackType = "report"   // 下行投递回执
	CallbackTypeUpstream CallbackType = "upstream" // 上行短信（用户回复）
)

// CallbackStatus 回调状态。
const (
	CallbackStatusDelivered = "delivered" // 已送达
	CallbackStatusFailed    = "failed"    // 失败
	CallbackStatusRejected  = "rejected"  // 被拒绝
	CallbackStatusUnknown   = "unknown"   // 未知
)

// CallbackLog 服务商回调日志（统一记录下行投递回执与上行短信）。
type CallbackLog struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	// RequestID 全链路追踪ID（按 push_log 关联回填）。
	RequestID         string       `json:"request_id" gorm:"size:64;index;comment:全链路追踪ID"`
	Type              CallbackType `json:"type" gorm:"size:20;not null;default:report;index;comment:回调类型 report/upstream"`
	TaskNo            string       `json:"task_no" gorm:"size:64;index;comment:任务编号"`
	AppID             uint         `json:"app_id" gorm:"index;comment:应用ID"`
	ProviderCode      string       `json:"provider_code" gorm:"size:50;not null;index;comment:服务商代码"`
	ProviderAccountID uint         `json:"provider_account_id" gorm:"index;comment:服务商账号ID"`
	ProviderID        string       `json:"provider_id" gorm:"size:64;index;comment:服务商消息ID(上行可空)"`
	Mobile            string       `json:"mobile" gorm:"size:32;index;comment:手机号(回执=接收方,上行=发送方)"`
	Content           string       `json:"content" gorm:"type:text;comment:上行短信回复内容"`
	CallbackStatus    string       `json:"callback_status" gorm:"size:20;comment:回调状态 delivered/failed/rejected"`
	ErrorCode         string       `json:"error_code" gorm:"size:64;comment:错误码"`
	ErrorMessage      string       `json:"error_message" gorm:"type:text;comment:错误信息"`
	RawData           string       `json:"raw_data" gorm:"type:text;comment:原始回调数据"`
	CreatedAt         time.Time    `json:"created_at" gorm:"index;comment:创建时间"`
}

// TableName 指定表名。
func (CallbackLog) TableName() string {
	return "msg_callback_logs"
}
