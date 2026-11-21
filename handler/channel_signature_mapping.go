package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/cast"
	"github.com/chihqiang/ten-go/rest"
)

// ChannelSignatureMappingHandler 通道-签名映射管理处理器。
type ChannelSignatureMappingHandler struct {
	svc   *svc.ServiceContext
	logic *logic.ChannelSignatureMappingLogic
}

// NewChannelSignatureMappingHandler 创建通道-签名映射管理处理器。
func NewChannelSignatureMappingHandler(s *svc.ServiceContext) *ChannelSignatureMappingHandler {
	return &ChannelSignatureMappingHandler{svc: s, logic: logic.NewChannelSignatureMappingLogic(s)}
}

// Create 创建签名映射。
func (h *ChannelSignatureMappingHandler) Create(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	var req dto.CreateChannelSignatureMappingRequest
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

// List 签名映射列表。
func (h *ChannelSignatureMappingHandler) List(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	var req dto.ChannelSignatureMappingListRequest
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

// Get 签名映射详情。
func (h *ChannelSignatureMappingHandler) Get(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	id := cast.To[uint](rest.PathParam(r, "mapping_id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid mapping id"))
		return
	}
	resp, err := h.logic.Get(r.Context(), channelID, id)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// Update 更新签名映射。
func (h *ChannelSignatureMappingHandler) Update(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	id := cast.To[uint](rest.PathParam(r, "mapping_id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid mapping id"))
		return
	}
	var req dto.UpdateChannelSignatureMappingRequest
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

// Delete 删除签名映射。
func (h *ChannelSignatureMappingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	id := cast.To[uint](rest.PathParam(r, "mapping_id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid mapping id"))
		return
	}
	if err := h.logic.Delete(r.Context(), channelID, id); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, nil)
}

// GetAvailableSignatures 获取可用签名（用于映射下拉）。
func (h *ChannelSignatureMappingHandler) GetAvailableSignatures(w http.ResponseWriter, r *http.Request) {
	list, err := h.logic.GetAvailableSignatures(r.Context())
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, list)
}
