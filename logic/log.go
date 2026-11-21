package logic

import (
	"context"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"
)

// LogLogic 日志查询逻辑（推送日志 / 回调日志）。
type LogLogic struct {
	svc *svc.ServiceContext
}

// NewLogLogic 创建日志查询逻辑。
func NewLogLogic(s *svc.ServiceContext) *LogLogic {
	return &LogLogic{svc: s}
}

// ListPushLogs 分页查询推送日志。
func (l *LogLogic) ListPushLogs(ctx context.Context, req *dto.LogListRequest) ([]model.PushLog, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.PushLog{})
	if req.TaskNo != "" {
		q = q.Where("task_no = ?", req.TaskNo)
	}
	if req.RequestID != "" {
		q = q.Where("request_id = ?", req.RequestID)
	}
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	if req.ProviderAccountID > 0 {
		q = q.Where("provider_account_id = ?", req.ProviderAccountID)
	}
	if req.Status != "" {
		q = q.Where("status = ?", req.Status)
	}
	if req.StartDate != "" {
		q = q.Where("created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		q = q.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.PushLog
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ListPushLogsByTask 按任务号查询推送日志（不分页）。
func (l *LogLogic) ListPushLogsByTask(ctx context.Context, taskNo string) ([]model.PushLog, error) {
	var logs []model.PushLog
	if err := l.svc.DB.WithContext(ctx).
		Where("task_no = ?", taskNo).
		Order("id DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// ListCallbacks 分页查询回调日志。
func (l *LogLogic) ListCallbacks(ctx context.Context, req *dto.CallbackListRequest) ([]model.CallbackLog, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.CallbackLog{})
	if req.Type != "" {
		q = q.Where("type = ?", req.Type)
	}
	if req.TaskNo != "" {
		q = q.Where("task_no = ?", req.TaskNo)
	}
	if req.RequestID != "" {
		q = q.Where("request_id = ?", req.RequestID)
	}
	if req.ProviderCode != "" {
		q = q.Where("provider_code = ?", req.ProviderCode)
	}
	if req.CallbackStatus != "" {
		q = q.Where("callback_status = ?", req.CallbackStatus)
	}
	if req.StartDate != "" {
		q = q.Where("created_at >= ?", req.StartDate+" 00:00:00")
	}
	if req.EndDate != "" {
		q = q.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var logs []model.CallbackLog
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ListCallbacksByTask 按任务号查询回调日志（不分页）。
func (l *LogLogic) ListCallbacksByTask(ctx context.Context, taskNo string) ([]model.CallbackLog, error) {
	var logs []model.CallbackLog
	if err := l.svc.DB.WithContext(ctx).
		Where("task_no = ?", taskNo).
		Order("id DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
