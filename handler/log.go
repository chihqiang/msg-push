package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/rest"
)

// LogHandler 日志查询处理器（推送日志 / 回调日志）。
type LogHandler struct {
	svc   *svc.ServiceContext
	logic *logic.LogLogic
}

// NewLogHandler 创建日志查询处理器。
func NewLogHandler(s *svc.ServiceContext) *LogHandler {
	return &LogHandler{svc: s, logic: logic.NewLogLogic(s)}
}

// ListPushLogs 推送日志列表。
func (h *LogHandler) ListPushLogs(w http.ResponseWriter, r *http.Request) {
	var req dto.LogListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.ListPushLogs(r.Context(), &req)
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

// ListPushLogsByTask 按任务号查询推送日志。
func (h *LogHandler) ListPushLogsByTask(w http.ResponseWriter, r *http.Request) {
	taskNo := r.PathValue("task_id")
	if taskNo == "" {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "task_id is required"))
		return
	}
	list, err := h.logic.ListPushLogsByTask(r.Context(), taskNo)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, list)
}

// ListCallbacks 回调日志列表。
func (h *LogHandler) ListCallbacks(w http.ResponseWriter, r *http.Request) {
	var req dto.CallbackListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.ListCallbacks(r.Context(), &req)
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

// ListCallbacksByTask 按任务号查询回调日志。
func (h *LogHandler) ListCallbacksByTask(w http.ResponseWriter, r *http.Request) {
	taskNo := r.PathValue("task_id")
	if taskNo == "" {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "task_id is required"))
		return
	}
	list, err := h.logic.ListCallbacksByTask(r.Context(), taskNo)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, list)
}
