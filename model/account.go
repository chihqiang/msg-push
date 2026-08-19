package model

import (
	"time"

	"gorm.io/gorm"
)

// Account 商户（平台商户主体），登录后自助管理其应用、通道、模板、服务商账号等资源。
type Account struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	Username  string         `json:"username" gorm:"size:64;uniqueIndex;not null;comment:用户名"`
	Password  string         `json:"-" gorm:"size:128;not null;comment:密码(bcrypt)"`
	Name      string         `json:"name" gorm:"size:64;comment:姓名"`
	Status    int8           `json:"status" gorm:"not null;default:1;comment:状态 1启用 0禁用"`
	CreatedAt time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (Account) TableName() string {
	return "msg_accounts"
}
