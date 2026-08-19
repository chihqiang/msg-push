package model

import "time"

// PushLog 推送日志，记录每次任务的投递结果与渠道回执。
type PushLog struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	// RequestID 全链路追踪ID（与 push_tasks.request_id 一致）。
	RequestID         string `json:"request_id" gorm:"size:64;index;comment:全链路追踪ID"`
	TaskID            uint   `json:"task_id" gorm:"index;comment:任务ID(关联push_tasks.id)"`
	TaskNo            string `json:"task_no" gorm:"size:64;index;comment:任务编号(task_id)"`
	AppID             uint   `json:"app_id" gorm:"index;comment:应用ID"`
	ProviderAccountID uint   `json:"provider_account_id" gorm:"index;comment:服务商账号ID"`
	ProviderMsgID     string `json:"provider_msg_id" gorm:"size:64;index;comment:服务商消息ID"`
	Receiver          string `json:"receiver" gorm:"size:128;not null;comment:接收方"`
	// IsTest 是否测试日志（测试任务不真实发送，模拟成功）。
	IsTest       bool      `json:"is_test" gorm:"not null;default:false;comment:是否测试日志(不真实发送)"`
	Status       string    `json:"status" gorm:"size:16;index;not null;comment:状态"`
	ProviderResp string    `json:"provider_resp" gorm:"type:text;comment:服务商响应"`
	ErrorCode    string    `json:"error_code" gorm:"size:64;comment:错误码"`
	ErrorMsg     string    `json:"error_msg" gorm:"size:512;comment:错误信息"`
	CostTime     int       `json:"cost_time" gorm:"comment:耗时(毫秒)"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:创建时间"`
}

// TableName 指定表名。
func (PushLog) TableName() string {
	return "msg_push_logs"
}
