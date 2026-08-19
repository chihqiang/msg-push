package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ProviderAccount 服务商账号配置。
type ProviderAccount struct {
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	AccountCode  string         `json:"account_code" gorm:"size:50;uniqueIndex;not null;comment:账号代码(唯一标识)"`
	AccountName  string         `json:"account_name" gorm:"size:100;not null;comment:账号名称"`
	ProviderCode string         `json:"provider_code" gorm:"size:50;not null;index;comment:服务商代码,如 aliyun_sms"`
	ProviderType string         `json:"provider_type" gorm:"size:20;not null;index;comment:消息类型 sms/email/wechat_work/dingtalk"`
	Config       string         `json:"config" gorm:"type:text;not null;comment:服务商配置(JSON)"`
	Status       int8           `json:"status" gorm:"not null;default:1;index;comment:状态 1启用 0禁用"`
	Remark       string         `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt    time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (ProviderAccount) TableName() string {
	return "msg_provider_accounts"
}

// GetConfig 反序列化配置。
func (p *ProviderAccount) GetConfig() (map[string]any, error) {
	cfg := map[string]any{}
	if p.Config == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(p.Config), &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// SetConfig 序列化配置。
func (p *ProviderAccount) SetConfig(cfg map[string]any) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	p.Config = string(data)
	return nil
}

// GetConfigInto 反序列化配置到指定结构体（带 json tag）。
func (p *ProviderAccount) GetConfigInto(dst any) error {
	if p.Config == "" {
		return nil
	}
	return json.Unmarshal([]byte(p.Config), dst)
}
