package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
	"gorm.io/gorm"
)

// WebhookLogic Webhook 配置管理与通知逻辑。
type WebhookLogic struct {
	svc *svc.ServiceContext
}

// NewWebhookLogic 创建 Webhook 配置管理逻辑。
func NewWebhookLogic(s *svc.ServiceContext) *WebhookLogic {
	return &WebhookLogic{svc: s}
}

// Create 创建 Webhook 配置。
func (l *WebhookLogic) Create(ctx context.Context, req *dto.CreateWebhookConfigRequest) (*model.WebhookConfig, error) {
	cfg := &model.WebhookConfig{
		Name:        req.Name,
		AppID:       req.AppID,
		WebhookURL:  req.WebhookURL,
		Secret:      req.Secret,
		Events:      req.Events,
		Status:      1,
		RetryCount:  3,
		Timeout:     5,
		Description: req.Description,
	}
	if cfg.Events == "" {
		cfg.Events = "success,failed"
	}
	if req.Status != nil {
		cfg.Status = *req.Status
	}
	if req.RetryCount != nil {
		cfg.RetryCount = *req.RetryCount
	}
	if req.Timeout != nil {
		cfg.Timeout = *req.Timeout
	}
	if err := l.svc.DB.WithContext(ctx).Create(cfg).Error; err != nil {
		return nil, err
	}
	return cfg, nil
}

// List 分页查询 Webhook 配置（按名称搜索）。
func (l *WebhookLogic) List(ctx context.Context, req *dto.WebhookConfigListRequest) ([]model.WebhookConfig, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.WebhookConfig{})
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if req.Key != "" {
		like := "%" + req.Key + "%"
		q = q.Where("name LIKE ? OR webhook_url LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var configs []model.WebhookConfig
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&configs).Error; err != nil {
		return nil, 0, err
	}
	return configs, total, nil
}

// Get 获取 Webhook 配置详情。
func (l *WebhookLogic) Get(ctx context.Context, id uint) (*model.WebhookConfig, error) {
	var cfg model.WebhookConfig
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("webhook config not found")
		}
		return nil, err
	}
	return &cfg, nil
}

// Update 更新 Webhook 配置。
func (l *WebhookLogic) Update(ctx context.Context, id uint, req *dto.UpdateWebhookConfigRequest) error {
	var cfg model.WebhookConfig
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&cfg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("webhook config not found")
		}
		return err
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AppID != nil {
		updates["app_id"] = *req.AppID
	}
	if req.WebhookURL != "" {
		updates["webhook_url"] = req.WebhookURL
	}
	if req.Secret != "" {
		updates["secret"] = req.Secret
	}
	if req.Events != "" {
		updates["events"] = req.Events
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.RetryCount != nil {
		updates["retry_count"] = *req.RetryCount
	}
	if req.Timeout != nil {
		updates["timeout"] = *req.Timeout
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(updates) == 0 {
		return nil
	}
	res := l.svc.DB.WithContext(ctx).Model(&model.WebhookConfig{}).
		Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("webhook config not found")
	}
	return nil
}

// Delete 删除 Webhook 配置。
func (l *WebhookLogic) Delete(ctx context.Context, id uint) error {
	res := l.svc.DB.WithContext(ctx).Delete(&model.WebhookConfig{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("webhook config not found")
	}
	return nil
}

// ListLogs 分页查询 Webhook 日志。
func (l *WebhookLogic) ListLogs(ctx context.Context, req *dto.WebhookLogListRequest) ([]model.WebhookLog, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.WebhookLog{})
	if req.TaskNo != "" {
		q = q.Where("task_no = ?", req.TaskNo)
	}
	if req.RequestID != "" {
		q = q.Where("request_id = ?", req.RequestID)
	}
	if req.Event != "" {
		q = q.Where("event = ?", req.Event)
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
	var logs []model.WebhookLog
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ListLogsByTask 按任务号查询 Webhook 日志（不分页）。
func (l *WebhookLogic) ListLogsByTask(ctx context.Context, taskNo string) ([]model.WebhookLog, error) {
	var logs []model.WebhookLog
	if err := l.svc.DB.WithContext(ctx).
		Where("task_no = ?", taskNo).
		Order("id DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// Dispatch 触发 Webhook 通知（outbox 模式）：查询启用的配置，订阅事件则写入待投递日志。
// 实际 HTTP 投递由 WebhookDispatcher 异步完成（含重试退避）。
// requestID 为全链路追踪ID（可为空），写入 webhook_log 并附带在通知 payload 中。
func (l *WebhookLogic) Dispatch(ctx context.Context, appID uint, taskNo, event string, payload map[string]any, requestID string) {
	// 追踪ID附带在通知 payload，便于接收方关联
	if requestID != "" {
		payload["request_id"] = requestID
	}
	body, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("webhook: marshal payload failed: %v", err)
		return
	}

	// 查询启用的、订阅了该事件的配置（匹配应用或全应用）
	var configs []model.WebhookConfig
	if err := l.svc.DB.WithContext(ctx).
		Where("status = 1").
		Find(&configs).Error; err != nil {
		logger.Errorf("webhook: query configs failed: %v", err)
		return
	}

	matched := false
	for i := range configs {
		cfg := &configs[i]
		if cfg.AppID != 0 && cfg.AppID != appID {
			continue
		}
		if !cfg.ShouldNotify(event) {
			continue
		}
		matched = true
		l.CreateOutbox(ctx, cfg, taskNo, appID, event, body, requestID)
	}
	if !matched {
		logger.Infof("webhook: no matching config for event=%s", event)
	}
}

// CreateOutbox 写入一条待投递的 webhook 日志（outbox），供 Dispatcher 异步投递。
func (l *WebhookLogic) CreateOutbox(ctx context.Context, cfg *model.WebhookConfig, taskNo string, appID uint, event string, body []byte, requestID string) {
	now := time.Now()
	log := &model.WebhookLog{
		RequestID:       requestID,
		TaskNo:          taskNo,
		AppID:           appID,
		WebhookConfigID: cfg.ID,
		WebhookURL:      cfg.WebhookURL,
		Event:           event,
		RequestData:     string(body),
		Status:          model.WebhookLogPending,
		MaxRetries:      cfg.RetryCount,
		TimeoutSeconds:  cfg.Timeout,
		SigningSecret:   cfg.Secret,
		NextAttemptAt:   now,
	}
	if log.MaxRetries <= 0 {
		log.MaxRetries = 3
	}
	if log.TimeoutSeconds <= 0 {
		log.TimeoutSeconds = 5
	}
	if err := l.svc.DB.WithContext(ctx).Create(log).Error; err != nil {
		logger.Errorf("webhook: create outbox failed: %v", err)
	}
}
