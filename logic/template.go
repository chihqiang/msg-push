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

// TemplateLogic 模板管理逻辑。
type TemplateLogic struct {
	svc *svc.ServiceContext
}

// NewTemplateLogic 创建模板管理逻辑。
func NewTemplateLogic(s *svc.ServiceContext) *TemplateLogic {
	return &TemplateLogic{svc: s}
}

// Create 创建模板。
func (l *TemplateLogic) Create(ctx context.Context, req *dto.CreateTemplateRequest) (*model.MessageTemplate, error) {
	// 校验通道存在
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", req.ChannelID).
		First(&model.Channel{}).Error; err != nil {
		return nil, errors.New("channel not found")
	}
	// 编码唯一性校验
	var count int64
	if err := l.svc.DB.WithContext(ctx).Model(&model.MessageTemplate{}).
		Where("code = ?", req.Code).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("template code already exists")
	}

	tpl := model.MessageTemplate{
		Code:      req.Code,
		ChannelID: req.ChannelID,
		Name:      req.Name,
		Content:   req.Content,
		Signature: req.Signature,
		Status:    1,
		Remark:    req.Remark,
	}
	if err := l.svc.DB.WithContext(ctx).Create(&tpl).Error; err != nil {
		return nil, err
	}
	return &tpl, nil
}

// List 分页查询模板列表，可按通道过滤、按编码/名称搜索。
func (l *TemplateLogic) List(ctx context.Context, page, pageSize int, channelID uint, key string) ([]model.MessageTemplate, int64, error) {
	var total int64
	var templates []model.MessageTemplate
	q := l.svc.DB.WithContext(ctx).Model(&model.MessageTemplate{})
	if channelID > 0 {
		q = q.Where("channel_id = ?", channelID)
	}
	if key != "" {
		like := "%" + key + "%"
		q = q.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}

// Get 查询模板详情。
func (l *TemplateLogic) Get(ctx context.Context, id uint) (*model.MessageTemplate, error) {
	var tpl model.MessageTemplate
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&tpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("template not found")
		}
		return nil, err
	}
	return &tpl, nil
}

// Update 更新模板。
func (l *TemplateLogic) Update(ctx context.Context, id uint, req *dto.UpdateTemplateRequest) error {
	updates := map[string]any{}
	if req.ChannelID > 0 {
		updates["channel_id"] = req.ChannelID
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Signature != "" {
		updates["signature"] = req.Signature
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
		Model(&model.MessageTemplate{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("template not found")
	}
	return nil
}

// Delete 删除模板（软删除）。
func (l *TemplateLogic) Delete(ctx context.Context, id uint) error {
	res := l.svc.DB.WithContext(ctx).Delete(&model.MessageTemplate{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("template not found")
	}
	return nil
}
