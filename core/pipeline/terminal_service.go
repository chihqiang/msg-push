// Package pipeline 消息处理管道：生产者入队 + 消费者投递，完成实际投递、状态流转与 webhook 触发。
package pipeline

import (
	"context"
	"time"

	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
	"gorm.io/gorm"
)

// TerminalTransition 终态流转输入。
type TerminalTransition struct {
	TaskID         string
	Status         string // success / failed
	Event          string // success / failed
	ProviderID     string
	ErrorCode      string
	ErrorMessage   string
	CallbackStatus string
	CallbackTime   *time.Time
}

// TerminalTransitionResult 流转结果。
type TerminalTransitionResult struct {
	Changed bool
}

// TerminalService 终态流转服务：CAS 更新任务终态 + 更新批次计数 + 触发 webhook。
type TerminalService struct {
	svc *svc.ServiceContext
	wh  *logic.WebhookLogic
}

// NewTerminalService 创建终态流转服务。
func NewTerminalService(s *svc.ServiceContext) *TerminalService {
	return &TerminalService{svc: s, wh: logic.NewWebhookLogic(s)}
}

// Transition 流转任务到终态（success/failed）。
// 幂等：仅当任务当前状态非终态时才生效（防回调与 worker 并发覆盖）。
func (s *TerminalService) Transition(ctx context.Context, tt TerminalTransition) (*TerminalTransitionResult, error) {
	if tt.TaskID == "" {
		return &TerminalTransitionResult{}, nil
	}

	updates := map[string]any{
		"status":     tt.Status,
		"updated_at": time.Now(),
	}
	if tt.ErrorMessage != "" {
		updates["error_msg"] = tt.ErrorMessage
	}
	if tt.CallbackStatus != "" {
		updates["callback_status"] = tt.CallbackStatus
	}
	if tt.CallbackTime != nil {
		updates["callback_time"] = *tt.CallbackTime
	}

	// CAS 更新：仅非终态可流转
	res := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("task_id = ? AND status NOT IN ?", tt.TaskID, []string{
			string(model.PushTaskStatusSuccess),
			string(model.PushTaskStatusFailed),
		}).
		Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return &TerminalTransitionResult{Changed: false}, nil
	}

	// 回读任务（用于批次计数与 webhook payload）
	var task model.PushTask
	if err := s.svc.DB.WithContext(ctx).Where("task_id = ?", tt.TaskID).First(&task).Error; err != nil {
		logger.Warnf("transition: reload task %s failed: %v", tt.TaskID, err)
		return &TerminalTransitionResult{Changed: true}, nil
	}

	// 更新批次计数
	s.updateBatchCount(ctx, &task, tt.Status)

	// 触发 webhook
	s.wh.Dispatch(ctx, task.AppID, task.TaskID, tt.Event, map[string]any{
		"task_id":       task.TaskID,
		"app_id":        task.AppID,
		"channel_id":    task.ChannelID,
		"receiver":      task.Receiver,
		"message_type":  task.MessageType,
		"status":        tt.Status,
		"provider_id":   tt.ProviderID,
		"error_code":    tt.ErrorCode,
		"error_message": tt.ErrorMessage,
		"occurred_at":   time.Now().Format(time.RFC3339),
	}, task.RequestID)

	return &TerminalTransitionResult{Changed: true}, nil
}

// updateBatchCount 更新批次计数（成功/失败，并在完成时置 completed）。
func (s *TerminalService) updateBatchCount(ctx context.Context, task *model.PushTask, status string) {
	if task.BatchID == "" {
		return
	}
	field := "success_count"
	if status == string(model.PushTaskStatusFailed) {
		field = "failed_count"
	}
	if err := s.svc.DB.WithContext(ctx).Model(&model.PushBatchTask{}).
		Where("batch_id = ?", task.BatchID).
		UpdateColumn(field, gorm.Expr(field+" + 1")).Error; err != nil {
		logger.Warnf("update batch %s count failed: %v", task.BatchID, err)
		return
	}
	// 若所有子任务已终态，批次置 completed
	s.syncBatchStatus(ctx, task)
}

// syncBatchStatus 同步批次状态与计数：统计所有子任务终态数量，更新批次计数并置 completed。
func (s *TerminalService) syncBatchStatus(ctx context.Context, task *model.PushTask) {
	var total, success, failed int64
	if err := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("batch_id = ?", task.BatchID).
		Count(&total).Error; err != nil {
		return
	}
	if err := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("batch_id = ? AND status = ?", task.BatchID,
			string(model.PushTaskStatusSuccess)).
		Count(&success).Error; err != nil {
		return
	}
	if err := s.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("batch_id = ? AND status = ?", task.BatchID,
			string(model.PushTaskStatusFailed)).
		Count(&failed).Error; err != nil {
		return
	}
	pending := total - success - failed
	if pending < 0 {
		pending = 0
	}
	status := string(model.PushBatchStatusProcessing)
	if pending == 0 && total > 0 {
		status = string(model.PushBatchStatusCompleted)
	}
	_ = s.svc.DB.WithContext(ctx).Model(&model.PushBatchTask{}).
		Where("batch_id = ?", task.BatchID).
		Updates(map[string]any{
			"status":        status,
			"success_count": success,
			"failed_count":  failed,
			"pending_count": pending,
			"updated_at":    time.Now(),
		}).Error
}
