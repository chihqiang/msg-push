package dto

import "time"

// ========== 通道-签名映射 ==========

// ChannelSignatureMappingListRequest 签名映射列表查询。
type ChannelSignatureMappingListRequest struct {
	Page     int `form:"page,default=1" binding:"omitempty,min=1"`
	PageSize int `form:"page_size,default=20" binding:"omitempty,min=1,max=100"`
}

// CreateChannelSignatureMappingRequest 创建签名映射。
type CreateChannelSignatureMappingRequest struct {
	SignatureName       string `json:"signature_name" binding:"required,max=100"`
	ProviderSignatureID uint   `json:"provider_signature_id" binding:"required"`
	ProviderID          uint   `json:"provider_id" binding:"required"`
	Status              *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// UpdateChannelSignatureMappingRequest 更新签名映射。
type UpdateChannelSignatureMappingRequest struct {
	SignatureName       string `json:"signature_name" binding:"omitempty,max=100"`
	ProviderSignatureID *uint  `json:"provider_signature_id"`
	ProviderID          *uint  `json:"provider_id"`
	Status              *int8  `json:"status" binding:"omitempty,oneof=0 1"`
}

// ChannelSignatureMappingResponse 签名映射响应。
type ChannelSignatureMappingResponse struct {
	ID                  uint      `json:"id"`
	ChannelID           uint      `json:"channel_id"`
	SignatureName       string    `json:"signature_name"`
	ProviderSignatureID uint      `json:"provider_signature_id"`
	SignatureCode       string    `json:"signature_code"`
	ProviderID          uint      `json:"provider_id"`
	ProviderName        string    `json:"provider_name"`
	ProviderType        string    `json:"provider_type"`
	Status              int8      `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
}

// AvailableProviderSignatureResponse 可用签名（用于映射下拉）。
type AvailableProviderSignatureResponse struct {
	ID            uint   `json:"id"`
	SignatureCode string `json:"signature_code"`
	SignatureName string `json:"signature_name"`
	ProviderID    uint   `json:"provider_id"`
	ProviderCode  string `json:"provider_code"`
	ProviderType  string `json:"provider_type"`
	Status        int8   `json:"status"`
}
