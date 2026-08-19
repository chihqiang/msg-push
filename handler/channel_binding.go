package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
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

// channelIDFromPath 解析通道 ID。
func channelIDFromPath(r *http.Request) (uint, bool) {
	channelID, err := parseID(r)
	if err != nil {
		return 0, false
	}
	return channelID, true
}

// Create 创建绑定。
func (h *ChannelBindingHandler) Create(w http.ResponseWriter, r *http.Request) {
	channelID, ok := channelIDFromPath(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	var req dto.CreateChannelBindingRequest
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

// List 绑定列表。
func (h *ChannelBindingHandler) List(w http.ResponseWriter, r *http.Request) {
	channelID, ok := channelIDFromPath(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	var req dto.ChannelBindingListRequest
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

// Get 绑定详情。
func (h *ChannelBindingHandler) Get(w http.ResponseWriter, r *http.Request) {
	channelID, ok := channelIDFromPath(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	id, err := parseIDValue(r, "binding_id")
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid binding id")
		return
	}
	resp, err := h.logic.Get(r.Context(), channelID, id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// Update 更新绑定。
func (h *ChannelBindingHandler) Update(w http.ResponseWriter, r *http.Request) {
	channelID, ok := channelIDFromPath(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	id, err := parseIDValue(r, "binding_id")
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid binding id")
		return
	}
	var req dto.UpdateChannelBindingRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), channelID, id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除绑定。
func (h *ChannelBindingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	channelID, ok := channelIDFromPath(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusBadRequest, "invalid channel id")
		return
	}
	id, err := parseIDValue(r, "binding_id")
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid binding id")
		return
	}
	if err := h.logic.Delete(r.Context(), channelID, id); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// GetAvailableTemplates 获取可用供应商模板（用于绑定下拉）。
func (h *ChannelBindingHandler) GetAvailableTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := h.logic.GetAvailableTemplates(r.Context())
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, list)
}
