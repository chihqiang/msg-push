package model

import (
	"regexp"
	"time"

	"gorm.io/gorm"
)

// 模板占位符正则：{key} 或 {{.key}}。
var templateVarRe = regexp.MustCompile(`\{\{?\.?([a-zA-Z0-9_]+)\}?\}`)

// MessageTemplate 消息模板，绑定通道，内容支持 {key} / {{.key}} 占位符渲染。
type MessageTemplate struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	// Code 模板唯一标识（业务编码，创建后不可修改，发送接口用它定位模板）。
	Code      string         `json:"code" gorm:"size:64;uniqueIndex;not null;comment:模板编码(唯一,不可修改)"`
	ChannelID uint           `json:"channel_id" gorm:"index;not null;comment:通道ID"`
	Name      string         `json:"name" gorm:"size:128;not null;comment:模板名称"`
	Content   string         `json:"content" gorm:"type:text;not null;comment:模板内容,支持{{.key}}"`
	Signature string         `json:"signature" gorm:"size:64;comment:签名/主题"`
	Status    int8           `json:"status" gorm:"not null;default:1;comment:状态 1启用 0禁用"`
	Remark    string         `json:"remark" gorm:"size:255;comment:备注"`
	CreatedAt time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (MessageTemplate) TableName() string {
	return "msg_message_templates"
}

// GetVariables 提取模板中的占位符变量名列表（去重、保序）。
func (m *MessageTemplate) GetVariables() []string {
	matches := templateVarRe.FindAllStringSubmatch(m.Content, -1)
	seen := map[string]bool{}
	vars := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		name := m[1]
		if !seen[name] {
			seen[name] = true
			vars = append(vars, name)
		}
	}
	return vars
}
