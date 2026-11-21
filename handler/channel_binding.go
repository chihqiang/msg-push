package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/cast"
	"github.com/chihqiang/ten-go/rest"
)

// ChannelBindingHandler 通道-模板绑定管理处理器。
type ChannelBindingHandler struct {
	svc   *svc.ServiceContext
	logic *logic.ChannelBindingLogic
}

// NewChannelBindingHandler 创建通道-模板绑定管理处理器。
func NewChannelBindingHandler(s *svc.ServiceContext) *ChannelBindingHandler {
	return &ChannelBindingHandler{svc: s, logic: logic.NewChannelBindingLogic(s)}
}

// Create 创建绑定。
func (h *ChannelBindingHandler) Create(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	var req dto.CreateChannelBindingRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	resp, err := h.logic.Create(r.Context(), channelID, &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// List 绑定列表。
func (h *ChannelBindingHandler) List(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	var req dto.ChannelBindingListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.List(r.Context(), channelID, &req)
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

// Get 绑定详情。
func (h *ChannelBindingHandler) Get(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	id := cast.To[uint](rest.PathParam(r, "binding_id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid binding id"))
		return
	}
	resp, err := h.logic.Get(r.Context(), channelID, id)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// Update 更新绑定。
func (h *ChannelBindingHandler) Update(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	id := cast.To[uint](rest.PathParam(r, "binding_id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid binding id"))
		return
	}
	var req dto.UpdateChannelBindingRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	if err := h.logic.Update(r.Context(), channelID, id, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除绑定。
func (h *ChannelBindingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	id := cast.To[uint](rest.PathParam(r, "binding_id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid binding id"))
		return
	}
	if err := h.logic.Delete(r.Context(), channelID, id); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, nil)
}

// GetAvailableTemplates 获取可用供应商模板（用于绑定下拉）。
func (h *ChannelBindingHandler) GetAvailableTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := h.logic.GetAvailableTemplates(r.Context())
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, list)
}
