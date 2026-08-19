package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// ParamMappingType 参数映射类型。
type ParamMappingType string

// 参数映射类型。
const (
	ParamMappingTypeFixed   ParamMappingType = "fixed"   // 固定值
	ParamMappingTypeMapping ParamMappingType = "mapping" // 映射系统变量
)

// ParamMappingItem 参数映射项。
type ParamMappingItem struct {
	Type        ParamMappingType `json:"type"`         // 映射类型 fixed/mapping
	ProviderVar string           `json:"provider_var"` // 供应商模板变量名
	SystemVar   string           `json:"system_var"`   // 系统变量名(type=mapping)
	Value       string           `json:"value"`        // 固定值(type=fixed)
}

// ChannelTemplateBinding 通道-供应商模板绑定。
// 同一通道下可绑定多个供应商模板，按优先级+权重分配流量。
type ChannelTemplateBinding struct {
	ID                   uint              `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	ChannelID            uint              `json:"channel_id" gorm:"not null;index;comment:通道ID"`
	ProviderTemplateID   uint              `json:"provider_template_id" gorm:"not null;comment:供应商模板ID"`
	ProviderID           uint              `json:"provider_id" gorm:"not null;index;comment:服务商账号ID(冗余)"`
	ParamMapping         string            `json:"-" gorm:"type:text;comment:参数映射(JSON)"`
	Weight               int               `json:"weight" gorm:"not null;default:10;comment:权重"`
	Priority             int               `json:"priority" gorm:"not null;default:100;comment:优先级(越小越优先)"`
	Status               int8              `json:"status" gorm:"not null;default:1;comment:状态 1启用 0禁用"`
	IsActive             int8              `json:"is_active" gorm:"not null;default:1;comment:是否激活 1是 0否"`
	AutoDisableOnFail    bool              `json:"auto_disable_on_fail" gorm:"not null;default:false;comment:失败时自动禁用"`
	AutoDisableThreshold int               `json:"auto_disable_threshold" gorm:"not null;default:5;comment:自动禁用阈值"`
	CreatedAt            time.Time         `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt            time.Time         `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt            gorm.DeletedAt    `json:"-" gorm:"index;comment:软删除时间"`
	ProviderTemplate     *ProviderTemplate `json:"provider_template,omitempty" gorm:"foreignKey:ProviderTemplateID;references:ID"`
	ProviderAccount      *ProviderAccount  `json:"provider_account,omitempty" gorm:"foreignKey:ProviderID;references:ID"`
}

// TableName 指定表名。
func (ChannelTemplateBinding) TableName() string {
	return "msg_channel_template_bindings"
}

// GetParamMapping 反序列化参数映射。
func (c *ChannelTemplateBinding) GetParamMapping() ([]ParamMappingItem, error) {
	var mapping []ParamMappingItem
	if c.ParamMapping == "" {
		return mapping, nil
	}
	if err := json.Unmarshal([]byte(c.ParamMapping), &mapping); err != nil {
		return nil, err
	}
	return mapping, nil
}

// SetParamMapping 序列化参数映射。
func (c *ChannelTemplateBinding) SetParamMapping(mapping []ParamMappingItem) error {
	data, err := json.Marshal(mapping)
	if err != nil {
		return err
	}
	c.ParamMapping = string(data)
	return nil
}
