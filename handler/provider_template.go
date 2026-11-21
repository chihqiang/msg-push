package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/cast"
	"github.com/chihqiang/ten-go/rest"
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
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	resp, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// List 供应商模板列表。
func (h *ProviderTemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.ProviderTemplateListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.List(r.Context(), &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, map[string]any{
		"list":      list,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// Get 供应商模板详情。
func (h *ProviderTemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	resp, err := h.logic.Get(r.Context(), id)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// Update 更新供应商模板。
func (h *ProviderTemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	var req dto.UpdateProviderTemplateRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除供应商模板。
func (h *ProviderTemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	if err := h.logic.Delete(r.Context(), id); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, nil)
}

// GetAvailableByProvider 获取指定服务商账号下的可用模板（用于通道绑定下拉）。
func (h *ProviderTemplateHandler) GetAvailableByProvider(w http.ResponseWriter, r *http.Request) {
	providerID := cast.To[uint](rest.PathParam(r, "id"))
	if providerID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid provider id"))
		return
	}
	list, err := h.logic.GetAvailableByProvider(r.Context(), providerID)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, list)
}
