package dto

import "time"

// CreateAppRequest 创建应用请求。
type CreateAppRequest struct {
	Name   string `json:"name" binding:"required,max=128"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
	// DailyQuota 每日发送配额，0=不限制，默认 10000。
	DailyQuota *int `json:"daily_quota" binding:"omitempty,min=0"`
	// RateLimit 每秒速率限制 QPS，默认 100。
	RateLimit *int `json:"rate_limit" binding:"omitempty,min=1"`
	// IsTest 是否测试应用：测试应用发送的消息不真实发送（模拟成功）。
	IsTest *bool `json:"is_test"`
}

// UpdateAppRequest 更新应用请求。
type UpdateAppRequest struct {
	Name   string `json:"name" binding:"omitempty,max=128"`
	Status *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Remark string `json:"remark" binding:"omitempty,max=255"`
	// DailyQuota 每日发送配额，0=不限制。
	DailyQuota *int `json:"daily_quota" binding:"omitempty,min=0"`
	// RateLimit 每秒速率限制 QPS。
	RateLimit *int `json:"rate_limit" binding:"omitempty,min=1"`
	// IsTest 是否测试应用。
	IsTest *bool `json:"is_test"`
}

// AppResponse 应用响应。Secret 仅在创建/重置密钥时返回明文。
type AppResponse struct {
	ID         uint      `json:"id"`
	AppID      string    `json:"app_id"`
	Name       string    `json:"name"`
	Secret     string    `json:"secret,omitempty"`
	Status     int8      `json:"status"`
	IsTest     bool      `json:"is_test"`
	DailyQuota int       `json:"daily_quota"`
	RateLimit  int       `json:"rate_limit"`
	Remark     string    `json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
}

// QuotaUsageResponse 应用配额使用情况。
type QuotaUsageResponse struct {
	DailyQuota      int     `json:"daily_quota"`
	TodayUsed       int     `json:"today_used"`
	Remaining       int     `json:"remaining"`
	UsagePercentage float64 `json:"usage_percentage"`
}
