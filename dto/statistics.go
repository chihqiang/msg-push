package dto

// TopListRequest 热门应用/近期活动查询。
type TopListRequest struct {
	Limit int `form:"limit,default=10" binding:"omitempty,min=1,max=50"`
}

// StatisticsRequest 统计查询请求。起止日期均为空时默认近 30 天。
type StatisticsRequest struct {
	StartDate   string `form:"start_date"` // YYYY-MM-DD
	EndDate     string `form:"end_date"`   // YYYY-MM-DD
	AppID       uint   `form:"app_id"`     // 应用ID（可选筛选）
	ChannelID   uint   `form:"channel_id"` // 通道ID（可选筛选）
	MessageType string `form:"message_type"`
}

// StatisticsCounts 统计中的通用任务状态计数。
type StatisticsCounts struct {
	TotalCount           int64    `json:"total_count"`
	SuccessCount         int64    `json:"success_count"`
	FailureCount         int64    `json:"failure_count"`
	PendingCount         int64    `json:"pending_count"`
	ProcessingCount      int64    `json:"processing_count"` // 发送中
	SentCount            int64    `json:"sent_count"`       // 已发送（当前无独立状态，恒 0）
	InProgressCount      int64    `json:"in_progress_count"`
	CompletedSuccessRate *float64 `json:"completed_success_rate"`
}

// StatisticsSummary 汇总统计。SuccessRate 保留 success / total 口径。
type StatisticsSummary struct {
	StatisticsCounts
	SuccessRate string `json:"success_rate"`
}

// DailyStatistics 每日统计。
type DailyStatistics struct {
	StatisticsSummary
	Date string `json:"date"`
}

// StatisticsPeriod 本次统计使用的时间范围和粒度。
type StatisticsPeriod struct {
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
	Granularity string `json:"granularity"`
}

// MessageTypeStatistics 消息类型维度统计。
type MessageTypeStatistics struct {
	StatisticsCounts
	MessageType string `json:"message_type"`
}

// ApplicationStatistics 应用维度 Top 统计。
type ApplicationStatistics struct {
	StatisticsCounts
	ID      uint   `json:"id"`
	AppID   string `json:"app_id"`
	AppName string `json:"app_name"`
}

// ChannelStatistics 通道维度 Top 统计。
type ChannelStatistics struct {
	StatisticsCounts
	ChannelID   uint   `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	ChannelType string `json:"channel_type"`
}

// StatisticsResponse 统计响应。
type StatisticsResponse struct {
	Period                  StatisticsPeriod         `json:"period"`
	Summary                 StatisticsSummary        `json:"summary"`
	Daily                   []*DailyStatistics       `json:"daily"`
	MessageTypeDistribution []*MessageTypeStatistics `json:"message_type_distribution"`
	TopApplications         []*ApplicationStatistics `json:"top_applications"`
	TopChannels             []*ChannelStatistics     `json:"top_channels"`
}

// DashboardResponse 仪表盘响应。
type DashboardResponse struct {
	TotalApplications         int64    `json:"total_applications"`
	ActiveApplications        int64    `json:"active_applications"`
	TotalChannels             int64    `json:"total_channels"`
	ActiveChannels            int64    `json:"active_channels"`
	TotalProviders            int64    `json:"total_providers"`
	ActiveProviders           int64    `json:"active_providers"`
	TodayPushCount            int64    `json:"today_push_count"`
	TodaySuccessCount         int64    `json:"today_success_count"`
	TodayFailedCount          int64    `json:"today_failed_count"`
	TodayInProgressCount      int64    `json:"today_in_progress_count"`
	TodaySuccessRate          string   `json:"today_success_rate"`
	TodayCompletedSuccessRate *float64 `json:"today_completed_success_rate"`
	TotalPushCount            int64    `json:"total_push_count"`
}

// TopApplicationResponse 热门应用。
type TopApplicationResponse struct {
	ID           uint   `json:"id"`
	AppID        string `json:"app_id"`
	AppName      string `json:"app_name"`
	PushCount    int64  `json:"push_count"`
	SuccessCount int64  `json:"success_count"`
	SuccessRate  string `json:"success_rate"`
}

// RecentActivityResponse 近期活动。
type RecentActivityResponse struct {
	ID          uint   `json:"id"`
	Description string `json:"description"`
	AppName     string `json:"app_name"`
	CreatedAt   string `json:"created_at"`
}
