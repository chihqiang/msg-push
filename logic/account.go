// Package logic 业务逻辑层：核心业务、校验与编排。
package logic

import (
	"context"
	"errors"
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/hash"
	"github.com/chihqiang/ten-go/rest"
	"gorm.io/gorm"
)

// AccountLogic 商户认证逻辑。
type AccountLogic struct {
	svc *svc.ServiceContext
}

// NewAccountLogic 创建商户认证逻辑。
func NewAccountLogic(s *svc.ServiceContext) *AccountLogic {
	return &AccountLogic{svc: s}
}

// Login 商户登录，校验商户账号密码并签发令牌对。
func (l *AccountLogic) Login(ctx context.Context, req *dto.AccountLoginRequest) (*dto.AccountLoginResponse, error) {
	var account model.Account
	err := l.svc.DB.WithContext(ctx).
		Where("username = ? AND status = 1", req.Username).
		First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, rest.NewCodeError(http.StatusUnauthorized, "invalid username or password")
	}
	if err != nil {
		return nil, err
	}
	if !hash.BcryptMatch(account.Password, req.Password) {
		return nil, rest.NewCodeError(http.StatusUnauthorized, "invalid username or password")
	}

	return l.issueToken(account)
}

// Refresh 使用刷新令牌换取新的令牌对。
func (l *AccountLogic) Refresh(ctx context.Context, refreshToken string) (*dto.AccountLoginResponse, error) {
	accountID, err := l.svc.Tokens.ParseRefresh(refreshToken)
	if err != nil {
		return nil, rest.NewCodeError(http.StatusUnauthorized, "invalid refresh token")
	}

	var account model.Account
	if err := l.svc.DB.WithContext(ctx).Where("id = ? AND status = 1", accountID).First(&account).Error; err != nil {
		return nil, rest.NewCodeError(http.StatusUnauthorized, "account not found or disabled")
	}

	return l.issueToken(account)
}

// issueToken 为商户签发令牌对。
func (l *AccountLogic) issueToken(account model.Account) (*dto.AccountLoginResponse, error) {
	pair, err := l.svc.Tokens.Issue(account.ID, account.Username)
	if err != nil {
		return nil, err
	}
	return &dto.AccountLoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt.Unix(),
	}, nil
}

// Profile 获取当前商户个人资料。
func (l *AccountLogic) Profile(ctx context.Context, accountID uint) (*dto.AccountProfileResponse, error) {
	var account model.Account
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", accountID).First(&account).Error; err != nil {
		return nil, err
	}
	return &dto.AccountProfileResponse{
		ID:        account.ID,
		Username:  account.Username,
		Name:      account.Name,
		Status:    account.Status,
		CreatedAt: account.CreatedAt,
	}, nil
}

// ChangePassword 修改当前商户密码（需校验旧密码）。
func (l *AccountLogic) ChangePassword(ctx context.Context, accountID uint, req *dto.ChangePasswordRequest) error {
	var account model.Account
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", accountID).First(&account).Error; err != nil {
		return err
	}
	if !hash.BcryptMatch(account.Password, req.OldPassword) {
		return rest.NewCodeError(http.StatusBadRequest, "旧密码不正确")
	}
	hashed, err := hash.BcryptHashDefault(req.NewPassword)
	if err != nil {
		return err
	}
	return l.svc.DB.WithContext(ctx).Model(&account).Update("password", hashed).Error
}
