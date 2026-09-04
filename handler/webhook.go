package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// WebhookHandler Webhook 配置管理处理器。
type WebhookHandler struct {
	svc   *svc.ServiceContext
	logic *logic.WebhookLogic
}

// NewWebhookHandler 创建 Webhook 配置管理处理器。
func NewWebhookHandler(s *svc.ServiceContext) *WebhookHandler {
	return &WebhookHandler{svc: s, logic: logic.NewWebhookLogic(s)}
}

// Create 创建 Webhook 配置。
func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWebhookConfigRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	cfg, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, cfg)
}

// List Webhook 配置列表。
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.WebhookConfigListRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.List(r.Context(), &req)
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

// Get Webhook 配置详情。
func (h *WebhookHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := httpx.PathValue(r, "id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	cfg, err := h.logic.Get(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, cfg)
}

// Update 更新 Webhook 配置。
func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := httpx.PathValue(r, "id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	var req dto.UpdateWebhookConfigRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除 Webhook 配置。
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// ListLogs Webhook 日志列表。
func (h *WebhookHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	var req dto.WebhookLogListRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.ListLogs(r.Context(), &req)
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

// ListLogsByTask 按任务号查询 Webhook 日志。
func (h *WebhookHandler) ListLogsByTask(w http.ResponseWriter, r *http.Request) {
	taskNo := r.PathValue("task_id")
	if taskNo == "" {
		writeBadRequest(r.Context(), w, "task_id is required")
		return
	}
	list, err := h.logic.ListLogsByTask(r.Context(), taskNo)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, list)
}
