package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
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
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	channel, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, channel)
}

// List 通道列表。
func (h *ChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.ChannelListRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	channels, total, err := h.logic.List(r.Context(), req.Page, req.PageSize, req.Type, req.Key)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, map[string]any{
		"list":      channels,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// Get 通道详情。
func (h *ChannelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := httpx.PathValue(r, "id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	channel, err := h.logic.Get(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, channel)
}

// Update 更新通道。
func (h *ChannelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := httpx.PathValue(r, "id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	var req dto.UpdateChannelRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除通道。
func (h *ChannelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := httpx.PathValue(r, "id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	if err := h.logic.Delete(r.Context(), id); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}
