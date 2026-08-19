package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// ChannelTestHandler 通道测试发送与健康历史处理器。
type ChannelTestHandler struct {
	svc   *svc.ServiceContext
	logic *logic.ChannelTestLogic
}

// NewChannelTestHandler 创建通道测试发送与健康历史处理器。
func NewChannelTestHandler(s *svc.ServiceContext) *ChannelTestHandler {
	return &ChannelTestHandler{svc: s, logic: logic.NewChannelTestLogic(s)}
}

// Test 通道测试发送。
func (h *ChannelTestHandler) Test(w http.ResponseWriter, r *http.Request) {
	channelID, ok := channelIDFromPath(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	var req dto.TestChannelRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	resp, err := h.logic.Test(r.Context(), channelID, &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// HealthHistory 通道健康历史。
func (h *ChannelTestHandler) HealthHistory(w http.ResponseWriter, r *http.Request) {
	channelID, ok := channelIDFromPath(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	var req dto.PageRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.HealthHistory(r.Context(), channelID, req.Page, req.PageSize)
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
