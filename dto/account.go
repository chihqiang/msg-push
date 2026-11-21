package dto

import "time"

// AccountLoginRequest 商户登录请求。
type AccountLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AccountLoginResponse 登录响应（令牌对）。
type AccountLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// AccountRefreshRequest 刷新令牌请求。
type AccountRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AccountProfileResponse 商户个人资料（当前登录账号）。
type AccountProfileResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ChangePasswordRequest 修改密码请求。
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
