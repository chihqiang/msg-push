package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
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
	channelID := httpx.PathValue(r, "id", uint(0))
	if channelID == 0 {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	var req dto.CreateChannelSignatureMappingRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	resp, err := h.logic.Create(r.Context(), channelID, &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// List 签名映射列表。
func (h *ChannelSignatureMappingHandler) List(w http.ResponseWriter, r *http.Request) {
	channelID := httpx.PathValue(r, "id", uint(0))
	if channelID == 0 {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	var req dto.ChannelSignatureMappingListRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.List(r.Context(), channelID, &req)
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

// Get 签名映射详情。
func (h *ChannelSignatureMappingHandler) Get(w http.ResponseWriter, r *http.Request) {
	channelID := httpx.PathValue(r, "id", uint(0))
	if channelID == 0 {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	id := httpx.PathValue(r, "mapping_id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid mapping id")
		return
	}
	resp, err := h.logic.Get(r.Context(), channelID, id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// Update 更新签名映射。
func (h *ChannelSignatureMappingHandler) Update(w http.ResponseWriter, r *http.Request) {
	channelID := httpx.PathValue(r, "id", uint(0))
	if channelID == 0 {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	id := httpx.PathValue(r, "mapping_id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid mapping id")
		return
	}
	var req dto.UpdateChannelSignatureMappingRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), channelID, id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除签名映射。
func (h *ChannelSignatureMappingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	channelID := httpx.PathValue(r, "id", uint(0))
	if channelID == 0 {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	id := httpx.PathValue(r, "mapping_id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid mapping id")
		return
	}
	if err := h.logic.Delete(r.Context(), channelID, id); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// GetAvailableSignatures 获取可用签名（用于映射下拉）。
func (h *ChannelSignatureMappingHandler) GetAvailableSignatures(w http.ResponseWriter, r *http.Request) {
	list, err := h.logic.GetAvailableSignatures(r.Context())
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, list)
}
