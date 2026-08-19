package model

import (
	"time"

	"gorm.io/gorm"
)

// ProviderSignature 服务商签名。
// SignatureCode 为实际发送用签名（短信签名/邮件主题），需与供应商平台报备一致。
type ProviderSignature struct {
	ID                uint             `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	ProviderAccountID uint             `json:"provider_account_id" gorm:"not null;index;comment:服务商账号ID"`
	SignatureCode     string           `json:"signature_code" gorm:"size:100;not null;comment:签名代码(实际发送用)"`
	SignatureName     string           `json:"signature_name" gorm:"size:100;not null;comment:签名名称(展示用)"`
	Status            int8             `json:"status" gorm:"not null;default:1;index;comment:状态 1启用 0禁用"`
	Remark            string           `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt         time.Time        `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt         time.Time        `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt         gorm.DeletedAt   `json:"-" gorm:"index;comment:软删除时间"`
	ProviderAccount   *ProviderAccount `json:"provider_account,omitempty" gorm:"foreignKey:ProviderAccountID;references:ID"`
}

// TableName 指定表名。
func (ProviderSignature) TableName() string {
	return "msg_provider_signatures"
}
