package model

import "time"

// WebhookLogStatus Webhook 投递状态。
type WebhookLogStatus string

// Webhook 投递状态常量。
const (
	WebhookLogPending    WebhookLogStatus = "pending"    // 待投递
	WebhookLogProcessing WebhookLogStatus = "processing" // 投递中（已认领）
	WebhookLogSuccess    WebhookLogStatus = "success"    // 成功
	WebhookLogFailed     WebhookLogStatus = "failed"     // 失败
)

// WebhookEvent Webhook 事件类型。
const (
	WebhookEventSuccess   = "success"   // 发送成功
	WebhookEventFailed    = "failed"    // 发送失败
	WebhookEventDelivered = "delivered" // 已送达（回执）
	WebhookEventUpstream  = "upstream"  // 上行短信
)

// WebhookLog Webhook 通知日志（outbox 模式）。
type WebhookLog struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	// RequestID 全链路追踪ID（与关联任务的 request_id 一致）。
	RequestID       string           `json:"request_id" gorm:"size:64;index;comment:全链路追踪ID"`
	TaskNo          string           `json:"task_no" gorm:"size:64;index;comment:任务编号"`
	AppID           uint             `json:"app_id" gorm:"index;comment:应用ID"`
	WebhookConfigID uint             `json:"webhook_config_id" gorm:"index;comment:Webhook配置ID"`
	WebhookURL      string           `json:"webhook_url" gorm:"size:500;not null;comment:Webhook地址"`
	Event           string           `json:"event" gorm:"size:20;not null;comment:事件类型"`
	RequestData     string           `json:"request_data" gorm:"type:text;comment:请求数据"`
	ResponseStatus  int              `json:"response_status" gorm:"comment:HTTP响应状态码"`
	ResponseData    string           `json:"response_data" gorm:"type:text;comment:响应内容"`
	Status          WebhookLogStatus `json:"status" gorm:"size:20;not null;index;comment:状态 pending/processing/success/failed"`
	ErrorMessage    string           `json:"error_message" gorm:"type:text;comment:错误信息"`
	RetryCount      int              `json:"retry_count" gorm:"not null;default:0;comment:已重试次数"`
	MaxRetries      int              `json:"max_retries" gorm:"not null;default:3;comment:最大重试次数"`
	TimeoutSeconds  int              `json:"timeout_seconds" gorm:"not null;default:5;comment:超时秒数"`
	// SigningSecret 签名密钥快照（来自 webhook_config），投递时用于 HMAC 签名。
	SigningSecret string `json:"-" gorm:"size:128;comment:签名密钥快照"`
	// NextAttemptAt 下次投递时间（重试退避用）。
	NextAttemptAt time.Time `json:"next_attempt_at" gorm:"index;comment:下次投递时间"`
	// LockedUntil 认领锁到期时间（processing 状态，NULL 表示未锁定）。
	LockedUntil *time.Time `json:"-" gorm:"comment:认领锁到期时间"`
	// LeaseToken 认领令牌（多实例互斥）。
	LeaseToken string    `json:"-" gorm:"size:64;comment:认领令牌"`
	CreatedAt  time.Time `json:"created_at" gorm:"index;comment:创建时间"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// TableName 指定表名。
func (WebhookLog) TableName() string {
	return "msg_webhook_logs"
}
