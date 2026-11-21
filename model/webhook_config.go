package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// WebhookConfig Webhook 通知配置，可绑定到具体应用（可选）。
type WebhookConfig struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	Name        string         `json:"name" gorm:"size:100;not null;comment:配置名称"`
	AppID       uint           `json:"app_id" gorm:"index;comment:应用ID(0=全应用)"`
	WebhookURL  string         `json:"webhook_url" gorm:"size:500;not null;comment:回调地址"`
	Secret      string         `json:"-" gorm:"size:128;comment:签名密钥"`
	Events      string         `json:"events" gorm:"size:200;not null;default:success,failed;comment:订阅事件 success,failed,delivered"`
	Status      int8           `json:"status" gorm:"not null;default:1;comment:状态 1启用 0禁用"`
	RetryCount  int            `json:"retry_count" gorm:"not null;default:3;comment:最大重试次数"`
	Timeout     int            `json:"timeout" gorm:"not null;default:5;comment:超时时间(秒)"`
	Description string         `json:"description" gorm:"size:255;comment:描述"`
	CreatedAt   time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (WebhookConfig) TableName() string {
	return "msg_webhook_configs"
}

// IsEnabled 是否启用。
func (w *WebhookConfig) IsEnabled() bool {
	return w.Status == 1
}

// ShouldNotify 是否应通知某事件。
func (w *WebhookConfig) ShouldNotify(event string) bool {
	if !w.IsEnabled() {
		return false
	}
	for _, e := range strings.Split(w.Events, ",") {
		if strings.TrimSpace(e) == event {
			return true
		}
	}
	return false
}
