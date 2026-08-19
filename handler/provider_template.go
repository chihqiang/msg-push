package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// ProviderTemplateHandler 供应商模板管理处理器。
type ProviderTemplateHandler struct {
	svc   *svc.ServiceContext
	logic *logic.ProviderTemplateLogic
}

// NewProviderTemplateHandler 创建供应商模板管理处理器。
func NewProviderTemplateHandler(s *svc.ServiceContext) *ProviderTemplateHandler {
	return &ProviderTemplateHandler{svc: s, logic: logic.NewProviderTemplateLogic(s)}
}

// Create 创建供应商模板。
func (h *ProviderTemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProviderTemplateRequest
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

// List 供应商模板列表。
func (h *ProviderTemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.ProviderTemplateListRequest
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

// Get 供应商模板详情。
func (h *ProviderTemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
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

// Update 更新供应商模板。
func (h *ProviderTemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	var req dto.UpdateProviderTemplateRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除供应商模板。
func (h *ProviderTemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// GetAvailableByProvider 获取指定服务商账号下的可用模板（用于通道绑定下拉）。
func (h *ProviderTemplateHandler) GetAvailableByProvider(w http.ResponseWriter, r *http.Request) {
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
