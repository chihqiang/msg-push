package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ProviderTemplate 供应商模板。
type ProviderTemplate struct {
	ID              uint             `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	ProviderID      uint             `json:"provider_id" gorm:"not null;index;comment:服务商账号ID"`
	TemplateCode    string           `json:"template_code" gorm:"size:100;not null;comment:供应商模板代码,如 SMS_123456"`
	TemplateName    string           `json:"template_name" gorm:"size:200;not null;comment:供应商模板名称"`
	ContentType     string           `json:"content_type" gorm:"size:20;not null;default:text;comment:内容类型 text/html/markdown"`
	TemplateContent string           `json:"template_content" gorm:"type:text;comment:供应商模板内容,如 验证码{code}"`
	Variables       string           `json:"-" gorm:"type:text;comment:供应商模板变量列表(JSON)"`
	Status          int8             `json:"status" gorm:"not null;default:1;index;comment:状态 1启用 0禁用"`
	Remark          string           `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt       time.Time        `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt       time.Time        `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt       gorm.DeletedAt   `json:"-" gorm:"index;comment:软删除时间"`
	ProviderAccount *ProviderAccount `json:"provider_account,omitempty" gorm:"foreignKey:ProviderID;references:ID"`
}

// TableName 指定表名。
func (ProviderTemplate) TableName() string {
	return "msg_provider_templates"
}

// GetVariables 反序列化变量列表。
func (p *ProviderTemplate) GetVariables() ([]string, error) {
	var vars []string
	if p.Variables == "" {
		return vars, nil
	}
	if err := json.Unmarshal([]byte(p.Variables), &vars); err != nil {
		return nil, err
	}
	return vars, nil
}

// SetVariables 序列化变量列表。
func (p *ProviderTemplate) SetVariables(vars []string) error {
	data, err := json.Marshal(vars)
	if err != nil {
		return err
	}
	p.Variables = string(data)
	return nil
}
