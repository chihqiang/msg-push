package logic

import (
	"context"
	"errors"
	"fmt"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"gorm.io/gorm"
)

// ChannelLogic 通道管理逻辑。
type ChannelLogic struct {
	svc *svc.ServiceContext
}

// NewChannelLogic 创建通道管理逻辑。
func NewChannelLogic(s *svc.ServiceContext) *ChannelLogic {
	return &ChannelLogic{svc: s}
}

// Create 创建通道。
func (l *ChannelLogic) Create(ctx context.Context, req *dto.CreateChannelRequest) (*model.Channel, error) {
	// 编码唯一性校验
	var count int64
	if err := l.svc.DB.WithContext(ctx).Model(&model.Channel{}).
		Where("code = ?", req.Code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("channel code already exists")
	}

	channel := model.Channel{
		Code:   req.Code,
		Name:   req.Name,
		Type:   model.ChannelType(req.Type),
		Config: req.Config,
		Status: 1,
		Remark: req.Remark,
	}
	if err := l.svc.DB.WithContext(ctx).Create(&channel).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}

// List 分页查询通道列表，可按类型过滤、按编码/名称搜索。
func (l *ChannelLogic) List(ctx context.Context, page, pageSize int, channelType, key string) ([]model.Channel, int64, error) {
	var total int64
	var channels []model.Channel
	q := l.svc.DB.WithContext(ctx).Model(&model.Channel{})
	if channelType != "" {
		q = q.Where("type = ?", channelType)
	}
	if key != "" {
		like := "%" + key + "%"
		q = q.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&channels).Error; err != nil {
		return nil, 0, err
	}
	return channels, total, nil
}

// Get 查询通道详情。
func (l *ChannelLogic) Get(ctx context.Context, id uint) (*model.Channel, error) {
	var channel model.Channel
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&channel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("channel not found")
		}
		return nil, err
	}
	return &channel, nil
}

// Update 更新通道。
func (l *ChannelLogic) Update(ctx context.Context, id uint, req *dto.UpdateChannelRequest) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Config != "" {
		updates["config"] = req.Config
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if len(updates) == 0 {
		return nil
	}
	res := l.svc.DB.WithContext(ctx).
		Model(&model.Channel{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("channel not found")
	}
	return nil
}

// Delete 删除通道（软删除）。
func (l *ChannelLogic) Delete(ctx context.Context, id uint) error {
	res := l.svc.DB.WithContext(ctx).Delete(&model.Channel{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("channel not found")
	}
	return nil
}
