package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"gorm.io/gorm"
)

// ProviderTemplateLogic 供应商模板管理逻辑。
type ProviderTemplateLogic struct {
	svc *svc.ServiceContext
}

// NewProviderTemplateLogic 创建供应商模板管理逻辑。
func NewProviderTemplateLogic(s *svc.ServiceContext) *ProviderTemplateLogic {
	return &ProviderTemplateLogic{svc: s}
}

// getProviderAccount 返回服务商账号。
func (l *ProviderTemplateLogic) getProviderAccount(ctx context.Context, providerID uint) (*model.ProviderAccount, error) {
	var acc model.ProviderAccount
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", providerID).
		First(&acc).Error; err != nil {
		return nil, errors.New("provider account not found")
	}
	return &acc, nil
}

// Create 创建供应商模板（模板归属服务商账号）。
func (l *ProviderTemplateLogic) Create(ctx context.Context, req *dto.CreateProviderTemplateRequest) (*dto.ProviderTemplateResponse, error) {
	if _, err := l.getProviderAccount(ctx, req.ProviderID); err != nil {
		return nil, err
	}

	tpl := &model.ProviderTemplate{
		ProviderID:      req.ProviderID,
		TemplateCode:    req.TemplateCode,
		TemplateName:    req.TemplateName,
		ContentType:     req.ContentType,
		TemplateContent: req.TemplateContent,
		Status:          1,
		Remark:          req.Remark,
	}
	if tpl.ContentType == "" {
		tpl.ContentType = "text"
	}
	if req.Status != nil {
		tpl.Status = *req.Status
	}
	if err := tpl.SetVariables(req.Variables); err != nil {
		return nil, err
	}
	if err := l.svc.DB.WithContext(ctx).Create(tpl).Error; err != nil {
		return nil, err
	}
	return l.toResponse(ctx, tpl)
}

// List 分页查询供应商模板（可按服务商账号过滤）。
func (l *ProviderTemplateLogic) List(ctx context.Context, req *dto.ProviderTemplateListRequest) ([]*dto.ProviderTemplateResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.ProviderTemplate{})
	if req.ProviderID != nil && *req.ProviderID > 0 {
		q = q.Where("provider_id = ?", *req.ProviderID)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if req.Key != "" {
		like := "%" + req.Key + "%"
		q = q.Where("template_name LIKE ? OR template_code LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var templates []model.ProviderTemplate
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*dto.ProviderTemplateResponse, 0, len(templates))
	for i := range templates {
		r, err := l.toResponse(ctx, &templates[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, nil
}

// Get 获取供应商模板详情。
func (l *ProviderTemplateLogic) Get(ctx context.Context, id uint) (*dto.ProviderTemplateResponse, error) {
	var tpl model.ProviderTemplate
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&tpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("provider template not found")
		}
		return nil, err
	}
	return l.toResponse(ctx, &tpl)
}

// Update 更新供应商模板。
func (l *ProviderTemplateLogic) Update(ctx context.Context, id uint, req *dto.UpdateProviderTemplateRequest) error {
	var tpl model.ProviderTemplate
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&tpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("provider template not found")
		}
		return err
	}

	updates := map[string]any{}
	if req.TemplateCode != "" {
		updates["template_code"] = req.TemplateCode
	}
	if req.TemplateName != "" {
		updates["template_name"] = req.TemplateName
	}
	if req.ContentType != "" {
		updates["content_type"] = req.ContentType
	}
	if req.TemplateContent != "" {
		updates["template_content"] = req.TemplateContent
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.Variables != nil {
		jsonVars, err := marshalStringSlice(req.Variables)
		if err != nil {
			return err
		}
		updates["variables"] = jsonVars
	}
	if len(updates) == 0 {
		return nil
	}
	return l.svc.DB.WithContext(ctx).Model(&model.ProviderTemplate{}).
		Where("id = ?", id).Updates(updates).Error
}

// Delete 删除供应商模板。
// 被通道-模板绑定引用时禁止删除，避免悬空引用导致选通道失败。
func (l *ProviderTemplateLogic) Delete(ctx context.Context, id uint) error {
	var tpl model.ProviderTemplate
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&tpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("provider template not found")
		}
		return err
	}

	// 配置层引用检查：存在则阻止删除，并列出具体引用类型
	var refs []string
	checkRef := func(m any, field, desc string) {
		var count int64
		if err := l.svc.DB.WithContext(ctx).Model(m).Where(field+" = ?", id).Count(&count).Error; err == nil && count > 0 {
			refs = append(refs, fmt.Sprintf("%s(%d)", desc, count))
		}
	}
	checkRef(&model.ChannelTemplateBinding{}, "provider_template_id", "通道-模板绑定")
	if len(refs) > 0 {
		return fmt.Errorf("供应商模板仍被引用，无法删除：%s。请先解除相关绑定", strings.Join(refs, "、"))
	}

	res := l.svc.DB.WithContext(ctx).Delete(&model.ProviderTemplate{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("provider template not found")
	}
	return nil
}

// GetAvailableByProvider 获取指定服务商账号下的可用模板（用于通道绑定下拉）。
func (l *ProviderTemplateLogic) GetAvailableByProvider(ctx context.Context, providerID uint) ([]*dto.AvailableProviderTemplateResponse, error) {
	var templates []model.ProviderTemplate
	if err := l.svc.DB.WithContext(ctx).
		Where("provider_id = ? AND status = 1", providerID).
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
func (l *ProviderTemplateLogic) toResponse(ctx context.Context, tpl *model.ProviderTemplate) (*dto.ProviderTemplateResponse, error) {
	vars, err := tpl.GetVariables()
	if err != nil {
		return nil, err
	}
	return &dto.ProviderTemplateResponse{
		ID:              tpl.ID,
		ProviderID:      tpl.ProviderID,
		TemplateCode:    tpl.TemplateCode,
		TemplateName:    tpl.TemplateName,
		ContentType:     tpl.ContentType,
		TemplateContent: tpl.TemplateContent,
		Variables:       vars,
		Status:          tpl.Status,
		Remark:          tpl.Remark,
		CreatedAt:       tpl.CreatedAt,
		UpdatedAt:       tpl.UpdatedAt,
	}, nil
}
