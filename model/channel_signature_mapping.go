package model

import (
	"time"

	"gorm.io/gorm"
)

// ChannelSignatureMapping 通道-签名映射。
// 将用户自定义签名名称映射到供应商签名。
type ChannelSignatureMapping struct {
	ID                  uint               `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	ChannelID           uint               `json:"channel_id" gorm:"not null;index;comment:通道ID"`
	SignatureName       string             `json:"signature_name" gorm:"size:100;not null;comment:用户自定义签名名称"`
	ProviderSignatureID uint               `json:"provider_signature_id" gorm:"not null;comment:供应商签名ID"`
	ProviderID          uint               `json:"provider_id" gorm:"not null;index;comment:服务商账号ID(冗余)"`
	Status              int8               `json:"status" gorm:"not null;default:1;comment:状态 1启用 0禁用"`
	CreatedAt           time.Time          `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt           time.Time          `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt           gorm.DeletedAt     `json:"-" gorm:"index;comment:软删除时间"`
	ProviderSignature   *ProviderSignature `json:"provider_signature,omitempty" gorm:"foreignKey:ProviderSignatureID;references:ID"`
	ProviderAccount     *ProviderAccount   `json:"provider_account,omitempty" gorm:"foreignKey:ProviderID;references:ID"`
}

// TableName 指定表名。
func (ChannelSignatureMapping) TableName() string {
	return "msg_channel_signature_mappings"
}
