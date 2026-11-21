package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/middleware"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/rest"
)

// MessageHandler 消息处理器。
type MessageHandler struct {
	svc   *svc.ServiceContext
	logic *logic.MessageLogic
}

// NewMessageHandler 创建消息处理器。
func NewMessageHandler(s *svc.ServiceContext) *MessageHandler {
	return &MessageHandler{svc: s, logic: logic.NewMessageLogic(s)}
}

// Send 发送单条消息。
func (h *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	app, ok := middleware.AppFromContext(r.Context())
	if !ok {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusUnauthorized, "missing app context"))
		return
	}
	var req dto.SendRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	resp, err := h.logic.Send(r.Context(), app, &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// BatchSend 批量发送消息。
func (h *MessageHandler) BatchSend(w http.ResponseWriter, r *http.Request) {
	app, ok := middleware.AppFromContext(r.Context())
	if !ok {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusUnauthorized, "missing app context"))
		return
	}
	var req dto.BatchSendRequest
	if err := rest.Bind(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	resp, err := h.logic.BatchSend(r.Context(), app, &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// QueryTask 查询任务状态。
func (h *MessageHandler) QueryTask(w http.ResponseWriter, r *http.Request) {
	app, ok := middleware.AppFromContext(r.Context())
	if !ok {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusUnauthorized, "missing app context"))
		return
	}
	taskID := r.PathValue("task_id")
	if taskID == "" {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "task_id is required"))
		return
	}
	task, err := h.logic.QueryTask(r.Context(), app, taskID)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, task)
}
