package logic

import (
	"context"
	"errors"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"gorm.io/gorm"
)

// TaskLogic 任务 / 批量任务查询逻辑。
type TaskLogic struct {
	svc *svc.ServiceContext
}

// NewTaskLogic 创建任务查询逻辑。
func NewTaskLogic(s *svc.ServiceContext) *TaskLogic {
	return &TaskLogic{svc: s}
}

// ListTasks 分页查询任务列表。
func (l *TaskLogic) ListTasks(ctx context.Context, req *dto.PushTaskListRequest) ([]*dto.PushTaskResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.PushTask{})
	if req.TaskNo != "" {
		q = q.Where("task_id = ?", req.TaskNo)
	}
	if req.RequestID != "" {
		q = q.Where("request_id = ?", req.RequestID)
	}
	if req.BatchID != "" {
		q = q.Where("batch_id = ?", req.BatchID)
	}
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
	}
	if req.ChannelID > 0 {
		q = q.Where("channel_id = ?", req.ChannelID)
	}
	if req.Receiver != "" {
		q = q.Where("receiver LIKE ?", "%"+req.Receiver+"%")
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
	var tasks []model.PushTask
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*dto.PushTaskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, l.toTaskResponse(ctx, &tasks[i]))
	}
	return out, total, nil
}

// GetTask 获取任务详情。
func (l *TaskLogic) GetTask(ctx context.Context, id uint) (*dto.PushTaskResponse, error) {
	var task model.PushTask
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, err
	}
	return l.toTaskResponse(ctx, &task), nil
}

// GetTaskByNo 按任务编号获取任务详情。
func (l *TaskLogic) GetTaskByNo(ctx context.Context, taskNo string) (*dto.PushTaskResponse, error) {
	var task model.PushTask
	if err := l.svc.DB.WithContext(ctx).Where("task_id = ?", taskNo).First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, err
	}
	return l.toTaskResponse(ctx, &task), nil
}

// ListBatchTasks 分页查询批次列表。
func (l *TaskLogic) ListBatchTasks(ctx context.Context, req *dto.PushBatchTaskListRequest) ([]*dto.PushBatchTaskResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.PushBatchTask{})
	if req.BatchID != "" {
		q = q.Where("batch_id LIKE ?", "%"+req.BatchID+"%")
	}
	if req.AppID > 0 {
		q = q.Where("app_id = ?", req.AppID)
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
	var batches []model.PushBatchTask
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&batches).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*dto.PushBatchTaskResponse, 0, len(batches))
	for i := range batches {
		out = append(out, l.toBatchResponse(&batches[i]))
	}
	return out, total, nil
}

// GetBatchTask 获取批次详情。
func (l *TaskLogic) GetBatchTask(ctx context.Context, id uint) (*dto.PushBatchTaskResponse, error) {
	var batch model.PushBatchTask
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("batch task not found")
		}
		return nil, err
	}
	return l.toBatchResponse(&batch), nil
}

// GetBatchByID 按批次 ID 获取批次。
func (l *TaskLogic) GetBatchByID(ctx context.Context, batchID string) (*dto.PushBatchTaskResponse, error) {
	var batch model.PushBatchTask
	if err := l.svc.DB.WithContext(ctx).Where("batch_id = ?", batchID).First(&batch).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("batch task not found")
		}
		return nil, err
	}
	return l.toBatchResponse(&batch), nil
}

// ListTasksByBatch 下钻：按批次 ID 分页查询批次下所有任务。
func (l *TaskLogic) ListTasksByBatch(ctx context.Context, batchID string, page, pageSize int) ([]*dto.PushTaskResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.PushTask{}).
		Where("batch_id = ?", batchID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tasks []model.PushTask
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*dto.PushTaskResponse, 0, len(tasks))
	for i := range tasks {
		out = append(out, l.toTaskResponse(ctx, &tasks[i]))
	}
	return out, total, nil
}

// toTaskResponse 任务模型转响应。
func (l *TaskLogic) toTaskResponse(ctx context.Context, task *model.PushTask) *dto.PushTaskResponse {
	channelName := ""
	if task.ChannelID > 0 {
		var ch model.Channel
		_ = l.svc.DB.WithContext(ctx).Where("id = ?", task.ChannelID).First(&ch).Error
		channelName = ch.Name
	}
	return &dto.PushTaskResponse{
		ID:          task.ID,
		TaskID:      task.TaskID,
		RequestID:   task.RequestID,
		AppID:       task.AppID,
		BatchID:     task.BatchID,
		ChannelID:   task.ChannelID,
		TemplateID:  task.TemplateID,
		Receiver:    task.Receiver,
		Params:      task.Params,
		Signature:   task.Signature,
		IsTest:      task.IsTest,
		Status:      string(task.Status),
		ErrorMsg:    task.ErrorMsg,
		ScheduledAt: task.ScheduledAt,
		SentAt:      task.SentAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		ChannelName: channelName,
	}
}

// toBatchResponse 批次模型转响应（含完成率）。
func (l *TaskLogic) toBatchResponse(batch *model.PushBatchTask) *dto.PushBatchTaskResponse {
	rate := 0.0
	if batch.TotalCount > 0 {
		rate = float64(batch.SuccessCount+batch.FailedCount) / float64(batch.TotalCount) * 100
	}
	return &dto.PushBatchTaskResponse{
		ID:             batch.ID,
		AppID:          batch.AppID,
		BatchID:        batch.BatchID,
		ChannelID:      batch.ChannelID,
		TemplateID:     batch.TemplateID,
		TotalCount:     batch.TotalCount,
		SuccessCount:   batch.SuccessCount,
		FailedCount:    batch.FailedCount,
		PendingCount:   batch.PendingCount,
		IsTest:         batch.IsTest,
		Status:         string(batch.Status),
		CreatedAt:      batch.CreatedAt,
		UpdatedAt:      batch.UpdatedAt,
		CompletionRate: rate,
	}
}
