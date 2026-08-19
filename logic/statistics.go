package logic

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"gorm.io/gorm"
)

// 统计常量。
const (
	statisticsDateLayout       = "2006-01-02"
	statisticsDefaultRangeDays = 30
	statisticsMaxRangeDays     = 90
	statisticsTopLimit         = 10
	statisticsTimezoneLabel    = "Asia/Shanghai"
)

// statisticsRange 统计时间范围。
type statisticsRange struct {
	start     time.Time
	end       time.Time
	startDate string
	endDate   string
}

// statisticsAggregateRow 聚合行。
type statisticsAggregateRow struct {
	TotalCount   int64
	SuccessCount int64
	FailureCount int64
	PendingCount int64
	SendCount    int64
}

// statisticsAggregateSelect 状态计数聚合 SQL（当前状态集：pending/sending/success/failed）。
const statisticsAggregateSelect = `
	COUNT(*) AS total_count,
	COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) AS success_count,
	COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) AS failure_count,
	COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) AS pending_count,
	COALESCE(SUM(CASE WHEN status = 'sending' THEN 1 ELSE 0 END), 0) AS send_count`

// statisticsDateBucketExpression 按业务日期分组 SQL 表达式（存储即本地业务时间）。
func statisticsDateBucketExpression(dialect string) string {
	switch dialect {
	case "mysql":
		return "DATE_FORMAT(created_at, '%Y-%m-%d')"
	case "postgres":
		return "TO_CHAR(created_at, 'YYYY-MM-DD')"
	case "sqlite":
		return "strftime('%Y-%m-%d', created_at)"
	default:
		return "DATE(created_at)"
	}
}

// StatisticLogic 统计分析逻辑：Dashboard / 趋势 / 排行 / 最近动态。
type StatisticLogic struct {
	svc *svc.ServiceContext
}

// NewStatisticLogic 创建统计分析逻辑。
func NewStatisticLogic(s *svc.ServiceContext) *StatisticLogic {
	return &StatisticLogic{svc: s}
}

// userTaskQuery 基础查询：全部推送任务。
func (l *StatisticLogic) userTaskQuery(ctx context.Context) *gorm.DB {
	return l.svc.DB.WithContext(ctx).Model(&model.PushTask{})
}

// resolveRange 解析统计时间范围，默认近 30 天，上限 90 天。
func resolveRange(req *dto.StatisticsRequest, now time.Time) (statisticsRange, error) {
	if req == nil {
		return statisticsRange{}, errors.New("request is required")
	}
	startDate, endDate := req.StartDate, req.EndDate
	switch {
	case startDate == "" && endDate == "":
		today := businessDayStart(now)
		startDate = today.AddDate(0, 0, -(statisticsDefaultRangeDays - 1)).Format(statisticsDateLayout)
		endDate = today.Format(statisticsDateLayout)
	case startDate == "" || endDate == "":
		return statisticsRange{}, errors.New("start_date and end_date must be provided together")
	}

	start, err := time.ParseInLocation(statisticsDateLayout, startDate, now.Location())
	if err != nil {
		return statisticsRange{}, fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := time.ParseInLocation(statisticsDateLayout, endDate, now.Location())
	if err != nil {
		return statisticsRange{}, fmt.Errorf("invalid end_date: %w", err)
	}
	// end 为含当天的半开区间上界
	end = end.AddDate(0, 0, 1)

	if end.Sub(start) > time.Duration(statisticsMaxRangeDays)*24*time.Hour {
		return statisticsRange{}, fmt.Errorf("date range cannot exceed %d days", statisticsMaxRangeDays)
	}
	return statisticsRange{start: start, end: end, startDate: startDate, endDate: endDate}, nil
}

// businessDayStart 返回今天零点（本地业务时区）。
func businessDayStart(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// completedSuccessRate 完成率（成功/(成功+失败)，无完成量时返回 nil）。
func completedSuccessRate(success, failure int64) *float64 {
	completed := success + failure
	if completed == 0 {
		return nil
	}
	rate := math.Round(float64(success)/float64(completed)*10000) / 100
	return &rate
}

// legacySuccessRate 兼容口径成功率字符串。
func legacySuccessRate(success, total int64) string {
	if total == 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", float64(success)/float64(total)*100)
}

// toCounts 聚合行 → DTO 计数。
func toCounts(row statisticsAggregateRow) dto.StatisticsCounts {
	return dto.StatisticsCounts{
		TotalCount:           row.TotalCount,
		SuccessCount:         row.SuccessCount,
		FailureCount:         row.FailureCount,
		PendingCount:         row.PendingCount,
		ProcessingCount:      row.SendCount,
		SentCount:            0,
		InProgressCount:      row.PendingCount + row.SendCount,
		CompletedSuccessRate: completedSuccessRate(row.SuccessCount, row.FailureCount),
	}
}

// GetStatistics 获取统计（趋势/汇总/消息类型分布/应用与通道 Top）。
func (l *StatisticLogic) GetStatistics(ctx context.Context, req *dto.StatisticsRequest) (*dto.StatisticsResponse, error) {
	now := time.Now()
	dateRange, err := resolveRange(req, now)
	if err != nil {
		return nil, err
	}
	dialect := l.svc.DB.Dialector.Name()

	query := l.userTaskQuery(ctx).
		Where("created_at >= ? AND created_at < ?", dateRange.start, dateRange.end)

	if req.AppID > 0 {
		query = query.Where("app_id = ?", req.AppID)
	}
	if req.ChannelID > 0 {
		query = query.Where("channel_id = ?", req.ChannelID)
	}
	if req.MessageType != "" {
		query = query.Where("message_type = ?", req.MessageType)
	}

	// 汇总
	var summaryRow statisticsAggregateRow
	if err := query.Session(&gorm.Session{}).
		Select(statisticsAggregateSelect).Scan(&summaryRow).Error; err != nil {
		return nil, fmt.Errorf("aggregate summary: %w", err)
	}

	// 每日
	dateExpression := statisticsDateBucketExpression(dialect)
	var dailyRows []struct {
		Date      string
		Aggregate statisticsAggregateRow `gorm:"embedded"`
	}
	if err := query.Session(&gorm.Session{}).
		Select(dateExpression + " AS date, " + statisticsAggregateSelect).
		Group(dateExpression).Order("date ASC").Scan(&dailyRows).Error; err != nil {
		return nil, fmt.Errorf("aggregate daily: %w", err)
	}

	// 消息类型分布
	var typeRows []struct {
		MessageType string
		Aggregate   statisticsAggregateRow `gorm:"embedded"`
	}
	if err := query.Session(&gorm.Session{}).
		Select("message_type, " + statisticsAggregateSelect).
		Group("message_type").Order("total_count DESC, message_type ASC").Scan(&typeRows).Error; err != nil {
		return nil, fmt.Errorf("aggregate message type: %w", err)
	}

	// Top 应用
	var appRows []struct {
		AppID     uint
		Aggregate statisticsAggregateRow `gorm:"embedded"`
	}
	if err := query.Session(&gorm.Session{}).
		Select("app_id, " + statisticsAggregateSelect).
		Group("app_id").Order("total_count DESC, app_id ASC").Limit(statisticsTopLimit).
		Scan(&appRows).Error; err != nil {
		return nil, fmt.Errorf("aggregate top apps: %w", err)
	}
	appNames := l.loadAppNames(ctx, appIDs(appRows))

	// Top 通道
	var chRows []struct {
		ChannelID uint
		Aggregate statisticsAggregateRow `gorm:"embedded"`
	}
	if err := query.Session(&gorm.Session{}).
		Select("channel_id, " + statisticsAggregateSelect).
		Group("channel_id").Order("total_count DESC, channel_id ASC").Limit(statisticsTopLimit).
		Scan(&chRows).Error; err != nil {
		return nil, fmt.Errorf("aggregate top channels: %w", err)
	}
	chNames := l.loadChannelNames(ctx, channelIDs(chRows))

	summaryCounts := toCounts(summaryRow)
	resp := &dto.StatisticsResponse{
		Period: dto.StatisticsPeriod{
			StartDate:   dateRange.startDate,
			EndDate:     dateRange.endDate,
			Timezone:    statisticsTimezoneLabel,
			Granularity: "day",
		},
		Summary: dto.StatisticsSummary{
			StatisticsCounts: summaryCounts,
			SuccessRate:      legacySuccessRate(summaryRow.SuccessCount, summaryRow.TotalCount),
		},
		Daily:                   make([]*dto.DailyStatistics, 0, statisticsMaxRangeDays),
		MessageTypeDistribution: make([]*dto.MessageTypeStatistics, 0, len(typeRows)),
		TopApplications:         make([]*dto.ApplicationStatistics, 0, len(appRows)),
		TopChannels:             make([]*dto.ChannelStatistics, 0, len(chRows)),
	}

	// 逐日补全（含 0 数据日期）
	dailyByDate := map[string]statisticsAggregateRow{}
	for _, row := range dailyRows {
		dailyByDate[row.Date] = row.Aggregate
	}
	for day := dateRange.start; day.Before(dateRange.end); day = day.AddDate(0, 0, 1) {
		date := day.Format(statisticsDateLayout)
		row := dailyByDate[date]
		resp.Daily = append(resp.Daily, &dto.DailyStatistics{
			Date: date,
			StatisticsSummary: dto.StatisticsSummary{
				StatisticsCounts: toCounts(row),
				SuccessRate:      legacySuccessRate(row.SuccessCount, row.TotalCount),
			},
		})
	}

	for _, row := range typeRows {
		resp.MessageTypeDistribution = append(resp.MessageTypeDistribution, &dto.MessageTypeStatistics{
			MessageType:      row.MessageType,
			StatisticsCounts: toCounts(row.Aggregate),
		})
	}
	for _, row := range appRows {
		item := &dto.ApplicationStatistics{
			ID:               row.AppID,
			AppID:            fmt.Sprintf("app_%d", row.AppID),
			AppName:          appNames[row.AppID],
			StatisticsCounts: toCounts(row.Aggregate),
		}
		resp.TopApplications = append(resp.TopApplications, item)
	}
	for _, row := range chRows {
		item := &dto.ChannelStatistics{
			ChannelID:        row.ChannelID,
			ChannelName:      chNames[row.ChannelID].name,
			ChannelType:      chNames[row.ChannelID].typ,
			StatisticsCounts: toCounts(row.Aggregate),
		}
		resp.TopChannels = append(resp.TopChannels, item)
	}
	return resp, nil
}

// GetDashboard 获取仪表盘数据。
func (l *StatisticLogic) GetDashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	resp := &dto.DashboardResponse{}

	// 应用统计
	if err := l.svc.DB.WithContext(ctx).Model(&model.Application{}).Count(&resp.TotalApplications).Error; err != nil {
		return nil, fmt.Errorf("count applications: %w", err)
	}
	if err := l.svc.DB.WithContext(ctx).Model(&model.Application{}).
		Where("status = 1").Count(&resp.ActiveApplications).Error; err != nil {
		return nil, fmt.Errorf("count active applications: %w", err)
	}

	// 通道统计
	if err := l.svc.DB.WithContext(ctx).Model(&model.Channel{}).Count(&resp.TotalChannels).Error; err != nil {
		return nil, fmt.Errorf("count channels: %w", err)
	}
	if err := l.svc.DB.WithContext(ctx).Model(&model.Channel{}).
		Where("status = 1").Count(&resp.ActiveChannels).Error; err != nil {
		return nil, fmt.Errorf("count active channels: %w", err)
	}

	// 服务商账号统计
	if err := l.svc.DB.WithContext(ctx).Model(&model.ProviderAccount{}).Count(&resp.TotalProviders).Error; err != nil {
		return nil, fmt.Errorf("count providers: %w", err)
	}
	if err := l.svc.DB.WithContext(ctx).Model(&model.ProviderAccount{}).
		Where("status = 1").Count(&resp.ActiveProviders).Error; err != nil {
		return nil, fmt.Errorf("count active providers: %w", err)
	}

	// 今日推送
	now := time.Now()
	todayStart := businessDayStart(now)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	var todayStats statisticsAggregateRow
	if err := l.userTaskQuery(ctx).
		Select(statisticsAggregateSelect).
		Where("created_at >= ? AND created_at < ?", todayStart, tomorrowStart).
		Scan(&todayStats).Error; err != nil {
		return nil, fmt.Errorf("aggregate today: %w", err)
	}
	resp.TodayPushCount = todayStats.TotalCount
	resp.TodaySuccessCount = todayStats.SuccessCount
	resp.TodayFailedCount = todayStats.FailureCount
	resp.TodayInProgressCount = todayStats.PendingCount + todayStats.SendCount
	resp.TodaySuccessRate = legacySuccessRate(todayStats.SuccessCount, todayStats.TotalCount)
	resp.TodayCompletedSuccessRate = completedSuccessRate(todayStats.SuccessCount, todayStats.FailureCount)

	// 总推送量
	if err := l.userTaskQuery(ctx).Count(&resp.TotalPushCount).Error; err != nil {
		return nil, fmt.Errorf("count total pushes: %w", err)
	}
	return resp, nil
}

// GetTopApplications 获取热门应用。
func (l *StatisticLogic) GetTopApplications(ctx context.Context, limit int) ([]*dto.TopApplicationResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = statisticsTopLimit
	}
	var results []struct {
		AppID        uint
		PushCount    int64
		SuccessCount int64
	}
	if err := l.userTaskQuery(ctx).
		Select("app_id, COUNT(*) AS push_count, COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) AS success_count").
		Group("app_id").Order("push_count DESC").Limit(limit).Scan(&results).Error; err != nil {
		return nil, err
	}
	appNames := l.loadAppNames(ctx, appIDsFromResults(results))

	items := make([]*dto.TopApplicationResponse, 0, len(results))
	for _, res := range results {
		successRate := "0.00%"
		if res.PushCount > 0 {
			successRate = fmt.Sprintf("%.2f%%", float64(res.SuccessCount)/float64(res.PushCount)*100)
		}
		items = append(items, &dto.TopApplicationResponse{
			ID:           res.AppID,
			AppID:        fmt.Sprintf("app_%d", res.AppID),
			AppName:      appNames[res.AppID],
			PushCount:    res.PushCount,
			SuccessCount: res.SuccessCount,
			SuccessRate:  successRate,
		})
	}
	return items, nil
}

// GetRecentActivities 获取近期活动。
func (l *StatisticLogic) GetRecentActivities(ctx context.Context, limit int) ([]*dto.RecentActivityResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = statisticsTopLimit
	}
	var tasks []*model.PushTask
	if err := l.userTaskQuery(ctx).Order("created_at DESC").Limit(limit).Find(&tasks).Error; err != nil {
		return nil, err
	}

	appMap := map[uint]string{}
	items := make([]*dto.RecentActivityResponse, 0, len(tasks))
	for _, task := range tasks {
		name, ok := appMap[task.AppID]
		if !ok {
			name = l.loadAppName(ctx, task.AppID)
			appMap[task.AppID] = name
		}
		items = append(items, &dto.RecentActivityResponse{
			ID:          task.ID,
			Description: fmt.Sprintf("推送消息 (TaskID: %s) 状态: %s", task.TaskID, task.Status),
			AppName:     name,
			CreatedAt:   task.CreatedAt.Format(time.RFC3339),
		})
	}
	return items, nil
}

// ---- 应用名 / 通道名加载辅助 ----

type channelName struct {
	name string
	typ  string
}

func (l *StatisticLogic) loadAppName(ctx context.Context, appID uint) string {
	var app model.Application
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", appID).First(&app).Error; err != nil {
		return "未知应用"
	}
	return app.Name
}

func (l *StatisticLogic) loadAppNames(ctx context.Context, ids []uint) map[uint]string {
	result := map[uint]string{}
	if len(ids) == 0 {
		return result
	}
	var apps []model.Application
	if err := l.svc.DB.WithContext(ctx).
		Where("id IN ?", ids).Find(&apps).Error; err != nil {
		return result
	}
	for _, app := range apps {
		result[app.ID] = app.Name
	}
	for _, id := range ids {
		if _, ok := result[id]; !ok {
			result[id] = "未知应用"
		}
	}
	return result
}

func (l *StatisticLogic) loadChannelNames(ctx context.Context, ids []uint) map[uint]channelName {
	result := map[uint]channelName{}
	if len(ids) == 0 {
		return result
	}
	var channels []model.Channel
	if err := l.svc.DB.WithContext(ctx).
		Where("id IN ?", ids).Find(&channels).Error; err != nil {
		return result
	}
	for _, ch := range channels {
		result[ch.ID] = channelName{name: ch.Name, typ: string(ch.Type)}
	}
	for _, id := range ids {
		if _, ok := result[id]; !ok {
			result[id] = channelName{name: fmt.Sprintf("已删除通道 (#%d)", id)}
		}
	}
	return result
}

func appIDs(rows []struct {
	AppID     uint
	Aggregate statisticsAggregateRow `gorm:"embedded"`
}) []uint {
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AppID)
	}
	return ids
}

func channelIDs(rows []struct {
	ChannelID uint
	Aggregate statisticsAggregateRow `gorm:"embedded"`
}) []uint {
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ChannelID)
	}
	return ids
}

func appIDsFromResults(rows []struct {
	AppID        uint
	PushCount    int64
	SuccessCount int64
}) []uint {
	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AppID)
	}
	return ids
}
