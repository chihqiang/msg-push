package scheduler

import (
	"context"
	"time"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
	"gorm.io/gorm/clause"
)

// QuotaSyncer 配额统计同步器：定时将 push_tasks 按应用/服务商聚合写入配额统计表。
// 多实例安全：Redis 分布式锁保证同一时刻仅一个实例执行；Upsert 幂等。
type QuotaSyncer struct {
	svc      *svc.ServiceContext
	interval time.Duration
	lockKey  string
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewQuotaSyncer 创建同步器。
func NewQuotaSyncer(s *svc.ServiceContext) *QuotaSyncer {
	return &QuotaSyncer{
		svc:      s,
		interval: 1 * time.Hour,
		lockKey:  "msgpush:lock:quota_syncer",
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start 启动同步器（后台协程，立即执行一次）。
func (s *QuotaSyncer) Start() {
	go s.run()
}

// Stop 优雅停止。
func (s *QuotaSyncer) Stop() {
	close(s.stopCh)
	<-s.doneCh
}

// StartService 适配 service.Starter。
func (s *QuotaSyncer) StartService() { s.Start() }

// run 主循环。
func (s *QuotaSyncer) run() {
	defer close(s.doneCh)
	s.syncOnce()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.syncOnce()
		case <-s.stopCh:
			return
		}
	}
}

// syncOnce 执行一次同步（带分布式锁）。
func (s *QuotaSyncer) syncOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 分布式锁：TTL 10 分钟（须覆盖同步耗时）
	ok, err := s.svc.Redis.Client().SetNX(ctx, s.lockKey, 1, 10*time.Minute).Result()
	if err != nil || !ok {
		logger.Infof("quota sync skipped: lock unavailable")
		return
	}
	defer s.svc.Redis.Client().Del(context.Background(), s.lockKey)

	now := time.Now()
	todayStart := businessDayStart(now)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	logger.Infof("starting quota sync...")
	s.syncAppQuota(ctx, todayStart, tomorrowStart)
	s.syncProviderQuota(ctx, todayStart, tomorrowStart)
	logger.Infof("quota sync done")
}

// syncAppQuota 按应用聚合写入 msg_app_quota_stats。
func (s *QuotaSyncer) syncAppQuota(ctx context.Context, start, end time.Time) {
	var apps []model.Application
	if err := s.svc.DB.WithContext(ctx).
		Where("deleted_at IS NULL").Find(&apps).Error; err != nil {
		logger.Errorf("quota sync: list apps failed: %v", err)
		return
	}

	for _, app := range apps {
		var row struct {
			Total   int64
			Success int64
			Failed  int64
		}
		err := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
			Select(`COUNT(*) AS total,
				COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) AS success,
				COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) AS failed`).
			Where("app_id = ? AND created_at >= ? AND created_at < ?", app.ID, start, end).
			Scan(&row).Error
		if err != nil {
			logger.Errorf("quota sync: aggregate app %d failed: %v", app.ID, err)
			continue
		}

		stat := &model.AppQuotaStat{
			AppID:        app.AppID,
			StatDate:     start,
			TotalCount:   int(row.Total),
			SuccessCount: int(row.Success),
			FailedCount:  int(row.Failed),
			UpdatedAt:    time.Now(),
		}
		if err := s.svc.DB.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "app_id"}, {Name: "stat_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"total_count", "success_count", "failed_count", "updated_at"}),
		}).Create(stat).Error; err != nil {
			logger.Errorf("quota sync: upsert app stat %s failed: %v", app.AppID, err)
		}
	}
}

// syncProviderQuota 按服务商账号聚合写入 msg_provider_quota_stats。
func (s *QuotaSyncer) syncProviderQuota(ctx context.Context, start, end time.Time) {
	// 按服务商账号聚合（provider_account_id>0 才有归属）
	type aggRow struct {
		ProviderAccountID uint
		Total             int64
		Success           int64
		Failed            int64
	}
	var rows []aggRow
	err := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Select(`provider_account_id,
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) AS success,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) AS failed`).
		Where("provider_account_id > 0 AND created_at >= ? AND created_at < ?", start, end).
		Group("provider_account_id").
		Scan(&rows).Error
	if err != nil {
		logger.Errorf("quota sync: aggregate provider failed: %v", err)
		return
	}

	// 写入聚合结果（无需账号归属）
	for _, r := range rows {
		stat := &model.ProviderQuotaStat{
			ProviderChannelID: r.ProviderAccountID,
			StatDate:          start,
			TotalCount:        int(r.Total),
			SuccessCount:      int(r.Success),
			FailedCount:       int(r.Failed),
			UpdatedAt:         time.Now(),
		}
		if err := s.svc.DB.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "provider_channel_id"}, {Name: "stat_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"total_count", "success_count", "failed_count", "updated_at"}),
		}).Create(stat).Error; err != nil {
			logger.Errorf("quota sync: upsert provider stat %d failed: %v", r.ProviderAccountID, err)
		}
	}
}

// businessDayStart 返回今天零点（本地时区）。
func businessDayStart(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
