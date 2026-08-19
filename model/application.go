package model

import (
	"time"

	"gorm.io/gorm"
)

// Application 接入应用，每个调用方一个应用，通过 app_id + app_secret 鉴权。
type Application struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	AppID     string `json:"app_id" gorm:"size:64;uniqueIndex;not null;comment:应用标识"`
	AppSecret string `json:"-" gorm:"size:128;not null;comment:应用密钥(bcrypt)"`
	// AppSecretPlain 应用密钥明文，用于 HMAC-SHA256 签名认证。
	AppSecretPlain string `json:"-" gorm:"size:128;not null;default:'';comment:应用密钥明文(用于HMAC签名)"`
	Name           string `json:"name" gorm:"size:128;not null;comment:应用名称"`
	Status         int8   `json:"status" gorm:"not null;default:1;comment:状态 1启用 0禁用"`
	// IsTest 是否测试应用：测试应用发送的消息走完整链路但不真实发送（模拟成功）。
	IsTest bool `json:"is_test" gorm:"not null;default:false;comment:是否测试应用(不真实发送)"`
	// DailyQuota 每日发送配额，0=不限制。
	DailyQuota int `json:"daily_quota" gorm:"not null;default:10000;comment:每日发送配额,0=不限制"`
	// RateLimit 每秒速率限制 QPS，默认 100。
	RateLimit int            `json:"rate_limit" gorm:"not null;default:100;comment:每秒速率限制QPS"`
	Remark    string         `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (Application) TableName() string {
	return "msg_applications"
}
