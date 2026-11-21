package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/cast"
	"github.com/chihqiang/ten-go/rest"
)

// TemplateHandler 模板管理处理器。
type TemplateHandler struct {
	svc   *svc.ServiceContext
	logic *logic.TemplateLogic
}

// NewTemplateHandler 创建模板管理处理器。
func NewTemplateHandler(s *svc.ServiceContext) *TemplateHandler {
	return &TemplateHandler{svc: s, logic: logic.NewTemplateLogic(s)}
}

// Create 创建模板。
func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTemplateRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	tpl, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, tpl)
}

// List 模板列表。
func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.TemplateListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	templates, total, err := h.logic.List(r.Context(), req.Page, req.PageSize, req.ChannelID, req.Key)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, map[string]any{
		"list":      templates,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// Get 模板详情。
func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	tpl, err := h.logic.Get(r.Context(), id)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, tpl)
}

// Update 更新模板。
func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	var req dto.UpdateTemplateRequest
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

// Delete 删除模板。
func (h *TemplateHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
