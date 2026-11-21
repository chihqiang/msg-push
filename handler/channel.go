package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/cast"
	"github.com/chihqiang/ten-go/rest"
)

// ChannelHandler 通道管理处理器。
type ChannelHandler struct {
	svc   *svc.ServiceContext
	logic *logic.ChannelLogic
}

// NewChannelHandler 创建通道管理处理器。
func NewChannelHandler(s *svc.ServiceContext) *ChannelHandler {
	return &ChannelHandler{svc: s, logic: logic.NewChannelLogic(s)}
}

// Create 创建通道。
func (h *ChannelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateChannelRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	channel, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, channel)
}

// List 通道列表。
func (h *ChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.ChannelListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	channels, total, err := h.logic.List(r.Context(), req.Page, req.PageSize, req.Type, req.Key)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, map[string]any{
		"list":      channels,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// Get 通道详情。
func (h *ChannelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	channel, err := h.logic.Get(r.Context(), id)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, channel)
}

// Update 更新通道。
func (h *ChannelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	var req dto.UpdateChannelRequest
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

// Delete 删除通道。
func (h *ChannelHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
