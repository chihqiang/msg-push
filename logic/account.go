// Package logic 业务逻辑层：核心业务、校验与编排。
package logic

import (
	"context"
	"errors"
	"strconv"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/hash"
	"github.com/chihqiang/infra-go/jwt"
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
		return nil, errors.New("invalid username or password")
	}
	if err != nil {
		return nil, err
	}
	if !hash.BcryptMatch(account.Password, req.Password) {
		return nil, errors.New("invalid username or password")
	}

	return l.issueToken(ctx, account)
}

// Refresh 使用刷新令牌换取新的令牌对。
func (l *AccountLogic) Refresh(ctx context.Context, refreshToken string) (*dto.AccountLoginResponse, error) {
	claims, err := l.svc.JWT.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	userID, _ := claims[jwt.ClaimKeyUserID].(string)
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, errors.New("invalid token subject")
	}

	var account model.Account
	if err := l.svc.DB.WithContext(ctx).Where("id = ? AND status = 1", id).First(&account).Error; err != nil {
		return nil, errors.New("account not found or disabled")
	}

	pair, err := l.svc.JWT.RefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return &dto.AccountLoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
	}, nil
}

// issueToken 为商户签发令牌对。
func (l *AccountLogic) issueToken(ctx context.Context, account model.Account) (*dto.AccountLoginResponse, error) {
	pair, err := l.svc.JWT.GenerateTokenPair(jwt.Claims{
		jwt.ClaimKeyUserID:   strconv.FormatUint(uint64(account.ID), 10),
		jwt.ClaimKeyUsername: account.Username,
		jwt.ClaimKeyRole:     "account",
	})
	if err != nil {
		return nil, err
	}
	return &dto.AccountLoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresAt:    pair.ExpiresAt,
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
		return errors.New("旧密码不正确")
	}
	hashed, err := hash.BcryptHashDefault(req.NewPassword)
	if err != nil {
		return err
	}
	return l.svc.DB.WithContext(ctx).Model(&account).Update("password", hashed).Error
}
