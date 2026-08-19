package model

import "time"

// AppQuotaStat 应用配额统计表（按日聚合）。统计由配额同步器（scheduler/quota_syncer）写入，HTTP 接口仅查询。
type AppQuotaStat struct {
	ID           uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	AppID        string    `json:"app_id" gorm:"size:64;not null;uniqueIndex:uk_app_date;comment:应用标识"`
	StatDate     time.Time `json:"stat_date" gorm:"type:date;not null;uniqueIndex:uk_app_date;index:idx_app_stat_date;comment:统计日期"`
	TotalCount   int       `json:"total_count" gorm:"not null;default:0;comment:总发送数"`
	SuccessCount int       `json:"success_count" gorm:"not null;default:0;comment:成功数"`
	FailedCount  int       `json:"failed_count" gorm:"not null;default:0;comment:失败数"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// TableName 指定表名。
func (AppQuotaStat) TableName() string {
	return "msg_app_quota_stats"
}
