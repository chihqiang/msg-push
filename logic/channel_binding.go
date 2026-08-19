package logic

import (
	"context"
	"errors"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"gorm.io/gorm"
)

// ChannelBindingLogic 通道-模板绑定管理逻辑。
type ChannelBindingLogic struct {
	svc *svc.ServiceContext
}

// NewChannelBindingLogic 创建通道-模板绑定管理逻辑。
func NewChannelBindingLogic(s *svc.ServiceContext) *ChannelBindingLogic {
	return &ChannelBindingLogic{svc: s}
}

// getChannel 返回通道。
func (l *ChannelBindingLogic) getChannel(ctx context.Context, channelID uint) (*model.Channel, error) {
	var ch model.Channel
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", channelID).
		First(&ch).Error; err != nil {
		return nil, errors.New("channel not found")
	}
	return &ch, nil
}

// Create 创建绑定。
func (l *ChannelBindingLogic) Create(ctx context.Context, channelID uint, req *dto.CreateChannelBindingRequest) (*dto.ChannelBindingResponse, error) {
	if _, err := l.getChannel(ctx, channelID); err != nil {
		return nil, err
	}
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", req.ProviderTemplateID).
		First(&model.ProviderTemplate{}).Error; err != nil {
		return nil, errors.New("provider template not found")
	}
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", req.ProviderID).
		First(&model.ProviderAccount{}).Error; err != nil {
		return nil, errors.New("provider account not found")
	}

	binding := &model.ChannelTemplateBinding{
		ChannelID:            channelID,
		ProviderTemplateID:   req.ProviderTemplateID,
		ProviderID:           req.ProviderID,
		Weight:               10,
		Priority:             100,
		Status:               1,
		IsActive:             1,
		AutoDisableThreshold: 5,
	}
	if req.Weight != nil {
		binding.Weight = *req.Weight
	}
	if req.Priority != nil {
		binding.Priority = *req.Priority
	}
	if req.Status != nil {
		binding.Status = *req.Status
	}
	if req.IsActive != nil {
		binding.IsActive = *req.IsActive
	}
	if req.AutoDisableOnFail != nil {
		binding.AutoDisableOnFail = *req.AutoDisableOnFail
	}
	if req.AutoDisableThreshold != nil {
		binding.AutoDisableThreshold = *req.AutoDisableThreshold
	}
	if err := binding.SetParamMapping(req.ParamMapping); err != nil {
		return nil, err
	}
	if err := l.svc.DB.WithContext(ctx).Create(binding).Error; err != nil {
		return nil, err
	}
	return l.toResponse(ctx, binding)
}

// List 分页查询通道绑定。
func (l *ChannelBindingLogic) List(ctx context.Context, channelID uint, req *dto.ChannelBindingListRequest) ([]*dto.ChannelBindingResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.ChannelTemplateBinding{}).
		Where("channel_id = ?", channelID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var bindings []model.ChannelTemplateBinding
	if err := q.Order("priority ASC, id ASC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&bindings).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*dto.ChannelBindingResponse, 0, len(bindings))
	for i := range bindings {
		r, err := l.toResponse(ctx, &bindings[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, nil
}

// Get 获取绑定详情。
func (l *ChannelBindingLogic) Get(ctx context.Context, channelID, id uint) (*dto.ChannelBindingResponse, error) {
	var binding model.ChannelTemplateBinding
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ? AND channel_id = ?", id, channelID).
		First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("binding not found")
		}
		return nil, err
	}
	return l.toResponse(ctx, &binding)
}

// Update 更新绑定。
func (l *ChannelBindingLogic) Update(ctx context.Context, channelID, id uint, req *dto.UpdateChannelBindingRequest) error {
	var binding model.ChannelTemplateBinding
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ? AND channel_id = ?", id, channelID).
		First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("binding not found")
		}
		return err
	}

	updates := map[string]any{}
	if req.ProviderTemplateID != nil {
		if err := l.svc.DB.WithContext(ctx).
			Where("id = ?", *req.ProviderTemplateID).
			First(&model.ProviderTemplate{}).Error; err != nil {
			return errors.New("provider template not found")
		}
		updates["provider_template_id"] = *req.ProviderTemplateID
	}
	if req.ProviderID != nil {
		if err := l.svc.DB.WithContext(ctx).
			Where("id = ?", *req.ProviderID).
			First(&model.ProviderAccount{}).Error; err != nil {
			return errors.New("provider account not found")
		}
		updates["provider_id"] = *req.ProviderID
	}
	if req.Weight != nil {
		updates["weight"] = *req.Weight
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.AutoDisableOnFail != nil {
		updates["auto_disable_on_fail"] = *req.AutoDisableOnFail
	}
	if req.AutoDisableThreshold != nil {
		updates["auto_disable_threshold"] = *req.AutoDisableThreshold
	}
	if req.ParamMapping != nil {
		mappingJSON, err := marshalParamMapping(req.ParamMapping)
		if err != nil {
			return err
		}
		updates["param_mapping"] = mappingJSON
	}
	if len(updates) == 0 {
		return nil
	}
	return l.svc.DB.WithContext(ctx).Model(&model.ChannelTemplateBinding{}).
		Where("id = ?", id).Updates(updates).Error
}

// Delete 删除绑定。
func (l *ChannelBindingLogic) Delete(ctx context.Context, channelID, id uint) error {
	res := l.svc.DB.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Delete(&model.ChannelTemplateBinding{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("binding not found")
	}
	return nil
}

// GetAvailableTemplates 获取可用供应商模板（用于绑定下拉）。
func (l *ChannelBindingLogic) GetAvailableTemplates(ctx context.Context) ([]*dto.AvailableProviderTemplateResponse, error) {
	var templates []model.ProviderTemplate
	if err := l.svc.DB.WithContext(ctx).
		Where("status = 1").
		Order("id DESC").Find(&templates).Error; err != nil {
		return nil, err
	}
	out := make([]*dto.AvailableProviderTemplateResponse, 0, len(templates))
	for i := range templates {
		vars, _ := templates[i].GetVariables()
		var acc model.ProviderAccount
		_ = l.svc.DB.WithContext(ctx).Where("id = ?", templates[i].ProviderID).First(&acc).Error
		out = append(out, &dto.AvailableProviderTemplateResponse{
			ID:              templates[i].ID,
			TemplateCode:    templates[i].TemplateCode,
			TemplateName:    templates[i].TemplateName,
			TemplateContent: templates[i].TemplateContent,
			Variables:       vars,
			ProviderID:      acc.ID,
			ProviderCode:    acc.ProviderCode,
			ProviderType:    acc.ProviderType,
			Status:          templates[i].Status,
		})
	}
	return out, nil
}

// toResponse 模型转响应。
func (l *ChannelBindingLogic) toResponse(ctx context.Context, binding *model.ChannelTemplateBinding) (*dto.ChannelBindingResponse, error) {
	mapping, err := binding.GetParamMapping()
	if err != nil {
		return nil, err
	}
	var tpl model.ProviderTemplate
	_ = l.svc.DB.WithContext(ctx).Where("id = ?", binding.ProviderTemplateID).First(&tpl).Error
	var acc model.ProviderAccount
	_ = l.svc.DB.WithContext(ctx).Where("id = ?", binding.ProviderID).First(&acc).Error

	return &dto.ChannelBindingResponse{
		ID:                   binding.ID,
		ChannelID:            binding.ChannelID,
		ProviderTemplateID:   binding.ProviderTemplateID,
		ProviderTemplateName: tpl.TemplateName,
		ProviderID:           binding.ProviderID,
		ProviderName:         acc.AccountName,
		ProviderType:         acc.ProviderType,
		ParamMapping:         mapping,
		Weight:               binding.Weight,
		Priority:             binding.Priority,
		Status:               binding.Status,
		IsActive:             binding.IsActive,
		AutoDisableOnFail:    binding.AutoDisableOnFail,
		AutoDisableThreshold: binding.AutoDisableThreshold,
		CreatedAt:            binding.CreatedAt,
	}, nil
}
