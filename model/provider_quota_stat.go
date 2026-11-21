package model

import "time"

// ProviderQuotaStat 服务商配额统计表（按日聚合）。统计由配额同步器（scheduler/quota_syncer）写入，HTTP 接口仅查询。
type ProviderQuotaStat struct {
	ID                uint      `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	ProviderChannelID uint      `json:"provider_channel_id" gorm:"not null;uniqueIndex:uk_channel_date;comment:服务商通道ID"`
	StatDate          time.Time `json:"stat_date" gorm:"type:date;not null;uniqueIndex:uk_channel_date;index:idx_provider_stat_date;comment:统计日期"`
	TotalCount        int       `json:"total_count" gorm:"not null;default:0;comment:总发送数"`
	SuccessCount      int       `json:"success_count" gorm:"not null;default:0;comment:成功数"`
	FailedCount       int       `json:"failed_count" gorm:"not null;default:0;comment:失败数"`
	CreatedAt         time.Time `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// TableName 指定表名。
func (ProviderQuotaStat) TableName() string {
	return "msg_provider_quota_stats"
}
