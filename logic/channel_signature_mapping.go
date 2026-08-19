package logic

import (
	"context"
	"errors"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"gorm.io/gorm"
)

// ChannelSignatureMappingLogic 通道-签名映射管理逻辑。
type ChannelSignatureMappingLogic struct {
	svc *svc.ServiceContext
}

// NewChannelSignatureMappingLogic 创建通道-签名映射管理逻辑。
func NewChannelSignatureMappingLogic(s *svc.ServiceContext) *ChannelSignatureMappingLogic {
	return &ChannelSignatureMappingLogic{svc: s}
}

// getChannel 返回通道。
func (l *ChannelSignatureMappingLogic) getChannel(ctx context.Context, channelID uint) (*model.Channel, error) {
	var ch model.Channel
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", channelID).
		First(&ch).Error; err != nil {
		return nil, errors.New("channel not found")
	}
	return &ch, nil
}

// Create 创建签名映射。
func (l *ChannelSignatureMappingLogic) Create(ctx context.Context, channelID uint, req *dto.CreateChannelSignatureMappingRequest) (*dto.ChannelSignatureMappingResponse, error) {
	if _, err := l.getChannel(ctx, channelID); err != nil {
		return nil, err
	}
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", req.ProviderSignatureID).
		First(&model.ProviderSignature{}).Error; err != nil {
		return nil, errors.New("provider signature not found")
	}
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", req.ProviderID).
		First(&model.ProviderAccount{}).Error; err != nil {
		return nil, errors.New("provider account not found")
	}

	mapping := &model.ChannelSignatureMapping{
		ChannelID:           channelID,
		SignatureName:       req.SignatureName,
		ProviderSignatureID: req.ProviderSignatureID,
		ProviderID:          req.ProviderID,
		Status:              1,
	}
	if req.Status != nil {
		mapping.Status = *req.Status
	}
	if err := l.svc.DB.WithContext(ctx).Create(mapping).Error; err != nil {
		return nil, err
	}
	return l.toResponse(ctx, mapping), nil
}

// List 分页查询签名映射。
func (l *ChannelSignatureMappingLogic) List(ctx context.Context, channelID uint, req *dto.ChannelSignatureMappingListRequest) ([]*dto.ChannelSignatureMappingResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.ChannelSignatureMapping{}).
		Where("channel_id = ?", channelID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var mappings []model.ChannelSignatureMapping
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&mappings).Error; err != nil {
		return nil, 0, err
	}

	out := make([]*dto.ChannelSignatureMappingResponse, 0, len(mappings))
	for i := range mappings {
		out = append(out, l.toResponse(ctx, &mappings[i]))
	}
	return out, total, nil
}

// Get 获取签名映射详情。
func (l *ChannelSignatureMappingLogic) Get(ctx context.Context, channelID, id uint) (*dto.ChannelSignatureMappingResponse, error) {
	var mapping model.ChannelSignatureMapping
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ? AND channel_id = ?", id, channelID).
		First(&mapping).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("signature mapping not found")
		}
		return nil, err
	}
	return l.toResponse(ctx, &mapping), nil
}

// Update 更新签名映射。
func (l *ChannelSignatureMappingLogic) Update(ctx context.Context, channelID, id uint, req *dto.UpdateChannelSignatureMappingRequest) error {
	var mapping model.ChannelSignatureMapping
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ? AND channel_id = ?", id, channelID).
		First(&mapping).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("signature mapping not found")
		}
		return err
	}

	updates := map[string]any{}
	if req.SignatureName != "" {
		updates["signature_name"] = req.SignatureName
	}
	if req.ProviderSignatureID != nil {
		if err := l.svc.DB.WithContext(ctx).
			Where("id = ?", *req.ProviderSignatureID).
			First(&model.ProviderSignature{}).Error; err != nil {
			return errors.New("provider signature not found")
		}
		updates["provider_signature_id"] = *req.ProviderSignatureID
	}
	if req.ProviderID != nil {
		if err := l.svc.DB.WithContext(ctx).
			Where("id = ?", *req.ProviderID).
			First(&model.ProviderAccount{}).Error; err != nil {
			return errors.New("provider account not found")
		}
		updates["provider_id"] = *req.ProviderID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		return nil
	}
	return l.svc.DB.WithContext(ctx).Model(&model.ChannelSignatureMapping{}).
		Where("id = ?", id).Updates(updates).Error
}

// Delete 删除签名映射。
func (l *ChannelSignatureMappingLogic) Delete(ctx context.Context, channelID, id uint) error {
	res := l.svc.DB.WithContext(ctx).
		Where("channel_id = ?", channelID).
		Delete(&model.ChannelSignatureMapping{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("signature mapping not found")
	}
	return nil
}

// GetAvailableSignatures 获取可用签名（用于映射下拉）。
func (l *ChannelSignatureMappingLogic) GetAvailableSignatures(ctx context.Context) ([]*dto.AvailableProviderSignatureResponse, error) {
	var sigs []model.ProviderSignature
	if err := l.svc.DB.WithContext(ctx).
		Where("status = 1").
		Order("id DESC").Find(&sigs).Error; err != nil {
		return nil, err
	}
	out := make([]*dto.AvailableProviderSignatureResponse, 0, len(sigs))
	for i := range sigs {
		var acc model.ProviderAccount
		_ = l.svc.DB.WithContext(ctx).Where("id = ?", sigs[i].ProviderAccountID).First(&acc).Error
		out = append(out, &dto.AvailableProviderSignatureResponse{
			ID:            sigs[i].ID,
			SignatureCode: sigs[i].SignatureCode,
			SignatureName: sigs[i].SignatureName,
			ProviderID:    acc.ID,
			ProviderCode:  acc.ProviderCode,
			ProviderType:  acc.ProviderType,
			Status:        sigs[i].Status,
		})
	}
	return out, nil
}

// toResponse 模型转响应。
func (l *ChannelSignatureMappingLogic) toResponse(ctx context.Context, mapping *model.ChannelSignatureMapping) *dto.ChannelSignatureMappingResponse {
	var sig model.ProviderSignature
	_ = l.svc.DB.WithContext(ctx).Where("id = ?", mapping.ProviderSignatureID).First(&sig).Error
	var acc model.ProviderAccount
	_ = l.svc.DB.WithContext(ctx).Where("id = ?", mapping.ProviderID).First(&acc).Error

	return &dto.ChannelSignatureMappingResponse{
		ID:                  mapping.ID,
		ChannelID:           mapping.ChannelID,
		SignatureName:       mapping.SignatureName,
		ProviderSignatureID: mapping.ProviderSignatureID,
		SignatureCode:       sig.SignatureCode,
		ProviderID:          mapping.ProviderID,
		ProviderName:        acc.AccountName,
		ProviderType:        acc.ProviderType,
		Status:              mapping.Status,
		CreatedAt:           mapping.CreatedAt,
	}
}
