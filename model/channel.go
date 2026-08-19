package model

import (
	"time"

	"gorm.io/gorm"
)

// ChannelType 通道类型。
type ChannelType string

// 支持的通道类型。
const (
	ChannelTypeSMS      ChannelType = "sms"      // 短信
	ChannelTypeEmail    ChannelType = "email"    // 邮件
	ChannelTypeWeCom    ChannelType = "wecom"    // 企业微信
	ChannelTypeDingTalk ChannelType = "dingtalk" // 钉钉
)

// Channel 通道配置。Config 为预留字段，实际发送配置取自服务商账号（ProviderAccount.Config）。
type Channel struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	// Code 通道唯一标识（业务编码，创建后不可修改，发送接口用它定位通道）。
	Code      string         `json:"code" gorm:"size:64;uniqueIndex;not null;comment:通道编码(唯一,不可修改)"`
	Name      string         `json:"name" gorm:"size:64;uniqueIndex;not null;comment:通道名称"`
	Type      ChannelType    `json:"type" gorm:"size:32;not null;index;comment:通道类型"`
	Config    string         `json:"config" gorm:"type:text;comment:通道配置(JSON)"`
	Status    int8           `json:"status" gorm:"not null;default:1;comment:状态 1启用 0禁用"`
	Remark    string         `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (Channel) TableName() string {
	return "msg_channels"
}
