package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// PushTaskStatus 推送任务状态。
type PushTaskStatus string

// 推送任务状态。
const (
	PushTaskStatusPending PushTaskStatus = "pending" // 待发送
	PushTaskStatusSending PushTaskStatus = "sending" // 发送中
	PushTaskStatusSuccess PushTaskStatus = "success" // 发送成功
	PushTaskStatusFailed  PushTaskStatus = "failed"  // 发送失败
)

// PushTask 推送任务。HTTP 接口层负责创建任务并入队，真实投递由 worker 包完成。
type PushTask struct {
	ID     uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	TaskID string `json:"task_id" gorm:"size:64;uniqueIndex;not null;comment:任务ID"`
	// RequestID 全链路追踪ID（来自 HTTP 请求 X-Request-Id，贯穿提交→投递→回执）。
	RequestID   string `json:"request_id" gorm:"size:64;index;comment:全链路追踪ID"`
	AppID       uint   `json:"app_id" gorm:"index;not null;default:0;comment:应用ID"`
	BatchID     string `json:"batch_id" gorm:"size:64;index;comment:批次ID(批量发送时)"`
	ChannelID   uint   `json:"channel_id" gorm:"index;not null;comment:通道ID"`
	TemplateID  uint   `json:"template_id" gorm:"index;comment:模板ID"`
	MessageType string `json:"message_type" gorm:"size:20;index;comment:消息类型(sms/email/wecom/dingtalk)"`
	Receiver    string `json:"receiver" gorm:"size:128;not null;comment:接收方"`
	Params      string `json:"params" gorm:"type:text;comment:模板参数(JSON)"`
	Signature   string `json:"signature" gorm:"size:64;comment:签名/主题"`
	// IsTest 是否测试任务：走完整链路但不调用真实发送（模拟成功）。
	IsTest   bool           `json:"is_test" gorm:"not null;default:false;comment:是否测试(不真实发送)"`
	Status   PushTaskStatus `json:"status" gorm:"size:16;index;not null;default:pending;comment:状态"`
	ErrorMsg string         `json:"error_msg" gorm:"size:512;comment:错误信息"`
	// ProviderAccountID 实际选中的服务商账号ID（消费端写入）。
	ProviderAccountID uint `json:"provider_account_id" gorm:"index;comment:服务商账号ID"`
	// RetryCount 已重试次数。
	RetryCount int `json:"retry_count" gorm:"not null;default:0;comment:已重试次数"`
	// MaxRetry 最大重试次数。
	MaxRetry int `json:"max_retry" gorm:"not null;default:3;comment:最大重试次数"`
	// ExcludeProviderIDs 需排除的服务商账号ID列表(JSON数组)，规则引擎切换供应商时写入。
	ExcludeProviderIDs string `json:"-" gorm:"type:text;comment:排除的服务商账号ID列表(JSON)"`
	// CallbackStatus 回调状态：空/success/failed/timeout，由回调处理与超时扫描器写入。
	CallbackStatus string         `json:"callback_status" gorm:"size:20;index;comment:回调状态"`
	CallbackTime   *time.Time     `json:"callback_time" gorm:"comment:回调时间"`
	ScheduledAt    *time.Time     `json:"scheduled_at" gorm:"index;comment:计划发送时间"`
	SentAt         *time.Time     `json:"sent_at" gorm:"comment:实际发送时间"`
	CreatedAt      time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (PushTask) TableName() string {
	return "msg_push_tasks"
}

// GetExcludeProviderIDs 解析排除的服务商账号ID列表。
func (t *PushTask) GetExcludeProviderIDs() []uint {
	var ids []uint
	if t.ExcludeProviderIDs == "" {
		return ids
	}
	if err := json.Unmarshal([]byte(t.ExcludeProviderIDs), &ids); err != nil {
		return nil
	}
	return ids
}

// SetExcludeProviderIDs 设置排除的服务商账号ID列表。
func (t *PushTask) SetExcludeProviderIDs(ids []uint) {
	if len(ids) == 0 {
		t.ExcludeProviderIDs = "[]"
		return
	}
	b, _ := json.Marshal(ids)
	t.ExcludeProviderIDs = string(b)
}

// AddExcludeProviderID 追加一个需排除的服务商账号ID。
func (t *PushTask) AddExcludeProviderID(id uint) {
	ids := t.GetExcludeProviderIDs()
	for _, v := range ids {
		if v == id {
			return
		}
	}
	ids = append(ids, id)
	t.SetExcludeProviderIDs(ids)
}
