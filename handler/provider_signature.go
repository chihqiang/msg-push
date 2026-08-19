package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// ProviderSignatureHandler 服务商签名管理处理器。
type ProviderSignatureHandler struct {
	svc   *svc.ServiceContext
	logic *logic.ProviderSignatureLogic
}

// NewProviderSignatureHandler 创建服务商签名管理处理器。
func NewProviderSignatureHandler(s *svc.ServiceContext) *ProviderSignatureHandler {
	return &ProviderSignatureHandler{svc: s, logic: logic.NewProviderSignatureLogic(s)}
}

// Create 创建签名。
func (h *ProviderSignatureHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProviderSignatureRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	resp, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// List 签名列表。
func (h *ProviderSignatureHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.ProviderSignatureListRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.List(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, map[string]any{
		"list":      list,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// Get 签名详情。
func (h *ProviderSignatureHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	resp, err := h.logic.Get(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// Update 更新签名。
func (h *ProviderSignatureHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	var req dto.UpdateProviderSignatureRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除签名。
func (h *ProviderSignatureHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	if err := h.logic.Delete(r.Context(), id); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// GetAvailableByProvider 获取指定服务商账号下的可用签名（用于通道签名映射下拉）。
func (h *ProviderSignatureHandler) GetAvailableByProvider(w http.ResponseWriter, r *http.Request) {
	providerID, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid provider id")
		return
	}
	list, err := h.logic.GetAvailableByProvider(r.Context(), providerID)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, list)
}
