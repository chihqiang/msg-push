package dto

import "time"

// ========== 服务商签名 ==========

// ProviderSignatureListRequest 签名列表查询。
type ProviderSignatureListRequest struct {
	ProviderAccountID uint   `form:"provider_account_id"`
	Status            *int8  `form:"status" binding:"omitempty,oneof=0 1"`
	Key               string `form:"key" binding:"omitempty,max=100"` // 关键字：签名名称/签名编码
	Page              int    `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize          int    `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}

// CreateProviderSignatureRequest 创建签名。
type CreateProviderSignatureRequest struct {
	ProviderAccountID uint   `json:"provider_account_id" binding:"required"`
	SignatureCode     string `json:"signature_code" binding:"required,max=100"`
	SignatureName     string `json:"signature_name" binding:"required,max=100"`
	Status            *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Remark            string `json:"remark" binding:"omitempty,max=255"`
}

// UpdateProviderSignatureRequest 更新签名。
type UpdateProviderSignatureRequest struct {
	SignatureCode string `json:"signature_code" binding:"omitempty,max=100"`
	SignatureName string `json:"signature_name" binding:"omitempty,max=100"`
	Status        *int8  `json:"status" binding:"omitempty,oneof=0 1"`
	Remark        string `json:"remark" binding:"omitempty,max=255"`
}

// ProviderSignatureResponse 签名响应。
type ProviderSignatureResponse struct {
	ID                uint      `json:"id"`
	ProviderAccountID uint      `json:"provider_account_id"`
	ProviderCode      string    `json:"provider_code,omitempty"`
	ProviderType      string    `json:"provider_type,omitempty"`
	SignatureCode     string    `json:"signature_code"`
	SignatureName     string    `json:"signature_name"`
	Status            int8      `json:"status"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
