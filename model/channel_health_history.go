package model

import "time"

// ChannelHealthStatus 通道健康状态。
type ChannelHealthStatus string

// 通道健康状态。
const (
	ChannelHealthHealthy   ChannelHealthStatus = "healthy"   // 健康
	ChannelHealthUnhealthy ChannelHealthStatus = "unhealthy" // 不健康
)

// ChannelHealthHistory 通道健康历史记录。
type ChannelHealthHistory struct {
	ID                uint                `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	ProviderChannelID uint                `json:"provider_channel_id" gorm:"not null;index;comment:服务商通道ID"`
	CheckTime         time.Time           `json:"check_time" gorm:"not null;index;comment:检查时间"`
	Status            ChannelHealthStatus `json:"status" gorm:"size:20;not null;comment:状态 healthy/unhealthy"`
	ResponseTime      int                 `json:"response_time" gorm:"comment:响应时间(毫秒)"`
	ErrorCount        int                 `json:"error_count" gorm:"not null;default:0;comment:滑窗错误数"`
	SuccessRate       float64             `json:"success_rate" gorm:"comment:成功率(%)"`
	IsAvailable       int8                `json:"is_available" gorm:"not null;default:1;comment:是否可用 1是 0否"`
	CreatedAt         time.Time           `json:"created_at" gorm:"comment:创建时间"`
}

// TableName 指定表名。
func (ChannelHealthHistory) TableName() string {
	return "msg_channel_health_history"
}
