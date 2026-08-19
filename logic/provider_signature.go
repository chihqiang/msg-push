package logic

import (
	"context"
	"errors"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"gorm.io/gorm"
)

// ProviderSignatureLogic 服务商签名管理逻辑。
type ProviderSignatureLogic struct {
	svc *svc.ServiceContext
}

// NewProviderSignatureLogic 创建服务商签名管理逻辑。
func NewProviderSignatureLogic(s *svc.ServiceContext) *ProviderSignatureLogic {
	return &ProviderSignatureLogic{svc: s}
}

// getProviderAccount 返回服务商账号。
func (l *ProviderSignatureLogic) getProviderAccount(ctx context.Context, providerAccountID uint) (*model.ProviderAccount, error) {
	var acc model.ProviderAccount
	if err := l.svc.DB.WithContext(ctx).
		Where("id = ?", providerAccountID).
		First(&acc).Error; err != nil {
		return nil, errors.New("provider account not found")
	}
	return &acc, nil
}

// Create 创建签名（签名归属服务商账号）。
func (l *ProviderSignatureLogic) Create(ctx context.Context, req *dto.CreateProviderSignatureRequest) (*dto.ProviderSignatureResponse, error) {
	acc, err := l.getProviderAccount(ctx, req.ProviderAccountID)
	if err != nil {
		return nil, err
	}

	sig := &model.ProviderSignature{
		ProviderAccountID: req.ProviderAccountID,
		SignatureCode:     req.SignatureCode,
		SignatureName:     req.SignatureName,
		Status:            1,
		Remark:            req.Remark,
	}
	if req.Status != nil {
		sig.Status = *req.Status
	}
	if err := l.svc.DB.WithContext(ctx).Create(sig).Error; err != nil {
		return nil, err
	}
	return l.toResponse(ctx, sig, acc), nil
}

// List 分页查询签名（可按服务商账号过滤）。
func (l *ProviderSignatureLogic) List(ctx context.Context, req *dto.ProviderSignatureListRequest) ([]*dto.ProviderSignatureResponse, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.ProviderSignature{})
	if req.ProviderAccountID > 0 {
		q = q.Where("provider_account_id = ?", req.ProviderAccountID)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if req.Key != "" {
		like := "%" + req.Key + "%"
		q = q.Where("signature_name LIKE ? OR signature_code LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var sigs []model.ProviderSignature
	if err := q.Order("id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&sigs).Error; err != nil {
		return nil, 0, err
	}

	// 预加载服务商账号信息
	out := make([]*dto.ProviderSignatureResponse, 0, len(sigs))
	for i := range sigs {
		var acc model.ProviderAccount
		_ = l.svc.DB.WithContext(ctx).Where("id = ?", sigs[i].ProviderAccountID).First(&acc).Error
		out = append(out, l.toResponse(ctx, &sigs[i], &acc))
	}
	return out, total, nil
}

// Get 获取签名详情。
func (l *ProviderSignatureLogic) Get(ctx context.Context, id uint) (*dto.ProviderSignatureResponse, error) {
	var sig model.ProviderSignature
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&sig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("provider signature not found")
		}
		return nil, err
	}
	var acc model.ProviderAccount
	_ = l.svc.DB.WithContext(ctx).Where("id = ?", sig.ProviderAccountID).First(&acc).Error
	return l.toResponse(ctx, &sig, &acc), nil
}

// Update 更新签名。
func (l *ProviderSignatureLogic) Update(ctx context.Context, id uint, req *dto.UpdateProviderSignatureRequest) error {
	var sig model.ProviderSignature
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&sig).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("provider signature not found")
		}
		return err
	}

	updates := map[string]any{}
	if req.SignatureCode != "" {
		updates["signature_code"] = req.SignatureCode
	}
	if req.SignatureName != "" {
		updates["signature_name"] = req.SignatureName
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
	return l.svc.DB.WithContext(ctx).Model(&model.ProviderSignature{}).
		Where("id = ?", id).Updates(updates).Error
}

// Delete 删除签名。
func (l *ProviderSignatureLogic) Delete(ctx context.Context, id uint) error {
	res := l.svc.DB.WithContext(ctx).Delete(&model.ProviderSignature{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("provider signature not found")
	}
	return nil
}

// GetAvailableByProvider 获取指定服务商账号下的可用签名（用于通道签名映射下拉）。
func (l *ProviderSignatureLogic) GetAvailableByProvider(ctx context.Context, providerAccountID uint) ([]*dto.AvailableProviderSignatureResponse, error) {
	var sigs []model.ProviderSignature
	if err := l.svc.DB.WithContext(ctx).
		Where("provider_account_id = ? AND status = 1", providerAccountID).
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
func (l *ProviderSignatureLogic) toResponse(ctx context.Context, sig *model.ProviderSignature, acc *model.ProviderAccount) *dto.ProviderSignatureResponse {
	return &dto.ProviderSignatureResponse{
		ID:                sig.ID,
		ProviderAccountID: sig.ProviderAccountID,
		ProviderCode:      acc.ProviderCode,
		ProviderType:      acc.ProviderType,
		SignatureCode:     sig.SignatureCode,
		SignatureName:     sig.SignatureName,
		Status:            sig.Status,
		Remark:            sig.Remark,
		CreatedAt:         sig.CreatedAt,
		UpdatedAt:         sig.UpdatedAt,
	}
}
