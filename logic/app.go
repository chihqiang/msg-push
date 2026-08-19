package logic

import (
	"context"
	"errors"
	"fmt"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/hash"
	"github.com/chihqiang/infra-go/stringx"
	"gorm.io/gorm"
)

// AppLogic 应用管理逻辑。
type AppLogic struct {
	svc *svc.ServiceContext
}

// NewAppLogic 创建应用管理逻辑。
func NewAppLogic(s *svc.ServiceContext) *AppLogic {
	return &AppLogic{svc: s}
}

// Create 创建应用，返回应用记录与明文密钥（仅此一次可见）。
func (l *AppLogic) Create(ctx context.Context, req *dto.CreateAppRequest) (*model.Application, string, error) {
	secret := stringx.Randn(32, stringx.RandTypeAll)
	secretHash, err := hash.BcryptHashDefault(secret)
	if err != nil {
		return nil, "", err
	}

	dailyQuota := 10000
	rateLimit := 100
	if req.DailyQuota != nil {
		dailyQuota = *req.DailyQuota
	}
	if req.RateLimit != nil {
		rateLimit = *req.RateLimit
	}
	isTest := false
	if req.IsTest != nil {
		isTest = *req.IsTest
	}

	app := model.Application{
		AppID:          "app_" + stringx.Randn(16, stringx.RandTypeLower),
		AppSecret:      secretHash,
		AppSecretPlain: secret,
		Name:           req.Name,
		Status:         1,
		IsTest:         isTest,
		DailyQuota:     dailyQuota,
		RateLimit:      rateLimit,
		Remark:         req.Remark,
	}
	if err := l.svc.DB.WithContext(ctx).Create(&app).Error; err != nil {
		return nil, "", err
	}
	return &app, secret, nil
}

// List 分页查询应用列表。
func (l *AppLogic) List(ctx context.Context, page, pageSize int) ([]model.Application, int64, error) {
	var total int64
	var apps []model.Application
	q := l.svc.DB.WithContext(ctx).Model(&model.Application{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

// Get 查询应用详情。
func (l *AppLogic) Get(ctx context.Context, id uint) (*model.Application, error) {
	var app model.Application
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("application not found")
		}
		return nil, err
	}
	return &app, nil
}

// Update 更新应用。
func (l *AppLogic) Update(ctx context.Context, id uint, req *dto.UpdateAppRequest) error {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.DailyQuota != nil {
		updates["daily_quota"] = *req.DailyQuota
	}
	if req.RateLimit != nil {
		updates["rate_limit"] = *req.RateLimit
	}
	if req.IsTest != nil {
		updates["is_test"] = *req.IsTest
	}
	if len(updates) == 0 {
		return nil
	}
	res := l.svc.DB.WithContext(ctx).
		Model(&model.Application{}).
		Where("id = ?", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("application not found")
	}
	return nil
}

// Delete 删除应用（软删除）。
func (l *AppLogic) Delete(ctx context.Context, id uint) error {
	res := l.svc.DB.WithContext(ctx).Delete(&model.Application{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("application not found")
	}
	return nil
}

// ResetSecret 重置应用密钥，返回新明文密钥。
func (l *AppLogic) ResetSecret(ctx context.Context, id uint) (string, error) {
	secret := stringx.Randn(32, stringx.RandTypeAll)
	secretHash, err := hash.BcryptHashDefault(secret)
	if err != nil {
		return "", err
	}
	res := l.svc.DB.WithContext(ctx).
		Model(&model.Application{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"app_secret":       secretHash,
			"app_secret_plain": secret,
		})
	if res.Error != nil {
		return "", res.Error
	}
	if res.RowsAffected == 0 {
		return "", fmt.Errorf("application not found")
	}
	return secret, nil
}
