package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/cast"
	"github.com/chihqiang/ten-go/rest"
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
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	var req dto.TestChannelRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	resp, err := h.logic.Test(r.Context(), channelID, &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// HealthHistory 通道健康历史。
func (h *ChannelTestHandler) HealthHistory(w http.ResponseWriter, r *http.Request) {
	channelID := cast.To[uint](rest.PathParam(r, "id"))
	if channelID == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid channel id"))
		return
	}
	var req dto.PageRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.HealthHistory(r.Context(), channelID, req.Page, req.PageSize)
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
