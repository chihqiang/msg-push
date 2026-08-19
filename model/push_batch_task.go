package model

import (
	"time"

	"gorm.io/gorm"
)

// PushBatchStatus 批量任务状态。
type PushBatchStatus string

// 批量任务状态。
const (
	PushBatchStatusProcessing PushBatchStatus = "processing" // 处理中
	PushBatchStatusCompleted  PushBatchStatus = "completed"  // 已完成
	PushBatchStatusFailed     PushBatchStatus = "failed"     // 失败
)

// PushBatchTask 批量任务（批次）。
// 批量发送时创建一条批次记录，记录整批的总数/成功/失败/待处理计数与状态。
type PushBatchTask struct {
	ID           uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	AppID        uint   `json:"app_id" gorm:"index;not null;default:0;comment:应用ID"`
	BatchID      string `json:"batch_id" gorm:"size:64;uniqueIndex;not null;comment:批次ID"`
	ChannelID    uint   `json:"channel_id" gorm:"index;not null;comment:通道ID"`
	TemplateID   uint   `json:"template_id" gorm:"index;comment:模板ID"`
	TotalCount   int    `json:"total_count" gorm:"not null;default:0;comment:总任务数"`
	SuccessCount int    `json:"success_count" gorm:"not null;default:0;comment:成功数"`
	FailedCount  int    `json:"failed_count" gorm:"not null;default:0;comment:失败数"`
	PendingCount int    `json:"pending_count" gorm:"not null;default:0;comment:待处理数"`
	// IsTest 是否测试批次：批次内任务走完整链路但不真实发送（模拟成功）。
	IsTest    bool            `json:"is_test" gorm:"not null;default:false;comment:是否测试批次(不真实发送)"`
	Status    PushBatchStatus `json:"status" gorm:"size:20;not null;default:processing;index;comment:状态 processing/completed/failed"`
	CreatedAt time.Time       `json:"created_at" gorm:"index;comment:创建时间"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (PushBatchTask) TableName() string {
	return "msg_push_batch_tasks"
}
