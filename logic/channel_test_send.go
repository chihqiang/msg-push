package logic

import (
	"context"
	"errors"
	"time"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/queue"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
	"github.com/chihqiang/infra-go/stringx"
	"gorm.io/gorm"
)

// ChannelTestLogic 通道测试发送与健康历史逻辑。
type ChannelTestLogic struct {
	svc *svc.ServiceContext
}

// NewChannelTestLogic 创建通道测试发送与健康历史逻辑。
func NewChannelTestLogic(s *svc.ServiceContext) *ChannelTestLogic {
	return &ChannelTestLogic{svc: s}
}

// Test 通道测试发送：创建一条测试任务并入队（真实投递由消费端完成）。
func (l *ChannelTestLogic) Test(ctx context.Context, channelID uint, req *dto.TestChannelRequest) (*dto.TestChannelResponse, error) {
	var ch model.Channel
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ? AND status = 1", channelID).
		First(&ch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("channel not found or disabled")
		}
		return nil, err
	}

	task := &model.PushTask{
		TaskID:    "test_" + stringx.RandId(),
		RequestID: httpx.RequestIDFromContext(ctx),
		ChannelID: channelID,
		Receiver:  req.Receiver,
		Status:    model.PushTaskStatusPending,
		Params:    `{"test":true}`,
	}
	if err := l.svc.DB.WithContext(ctx).Create(task).Error; err != nil {
		return nil, err
	}

	if _, err := queue.EnqueueSendMessage(ctx, l.svc.Producer, queue.SendMessagePayload{
		TaskID:    task.TaskID,
		RequestID: task.RequestID,
		ChannelID: channelID,
		Receiver:  task.Receiver,
		Params:    map[string]string{"test": "true", "content": req.Content},
	}, nil); err != nil {
		_ = l.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
			Where("task_id = ?", task.TaskID).
			Updates(map[string]any{
				"status":    model.PushTaskStatusFailed,
				"error_msg": "enqueue failed: " + err.Error(),
			}).Error
		return nil, err
	}

	// 记录健康历史（入队成功视为通道可用）
	l.recordHealth(ctx, channelID, true, 0, "test enqueued")

	return &dto.TestChannelResponse{
		TaskID: task.TaskID,
		Status: string(task.Status),
	}, nil
}

// HealthHistory 分页查询通道健康历史。
func (l *ChannelTestLogic) HealthHistory(ctx context.Context, channelID uint, page, pageSize int) ([]*dto.ChannelHealthHistoryResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.ChannelHealthHistory{}).
		Where("provider_channel_id = ?", channelID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var records []model.ChannelHealthHistory
	if err := q.Order("check_time DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*dto.ChannelHealthHistoryResponse, 0, len(records))
	for i := range records {
		out = append(out, &dto.ChannelHealthHistoryResponse{
			ID:                records[i].ID,
			ProviderChannelID: records[i].ProviderChannelID,
			CheckTime:         records[i].CheckTime,
			Status:            records[i].Status,
			ResponseTime:      records[i].ResponseTime,
			ErrorCount:        records[i].ErrorCount,
			SuccessRate:       records[i].SuccessRate,
			IsAvailable:       records[i].IsAvailable,
		})
	}
	return out, total, nil
}

// recordHealth 记录一条健康历史记录。
func (l *ChannelTestLogic) recordHealth(ctx context.Context, channelID uint, ok bool, responseTime int, msg string) {
	status := model.ChannelHealthHealthy
	isAvailable := int8(1)
	if !ok {
		status = model.ChannelHealthUnhealthy
		isAvailable = 0
	}
	record := &model.ChannelHealthHistory{
		ProviderChannelID: channelID,
		CheckTime:         time.Now(),
		Status:            status,
		ResponseTime:      responseTime,
		ErrorCount:        0,
		SuccessRate:       100,
		IsAvailable:       isAvailable,
	}
	if !ok {
		record.SuccessRate = 0
		record.ErrorCount = 1
	}
	if err := l.svc.DB.WithContext(ctx).Create(record).Error; err != nil {
		logger.Errorf("record channel health failed: %v", err)
	}
}
