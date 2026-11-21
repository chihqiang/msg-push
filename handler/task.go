package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/core/cast"
	"github.com/chihqiang/ten-go/rest"
)

// TaskHandler 任务 / 批量任务查询处理器。
type TaskHandler struct {
	svc   *svc.ServiceContext
	logic *logic.TaskLogic
}

// NewTaskHandler 创建任务查询处理器。
func NewTaskHandler(s *svc.ServiceContext) *TaskHandler {
	return &TaskHandler{svc: s, logic: logic.NewTaskLogic(s)}
}

// ListTasks 任务列表。
func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	var req dto.PushTaskListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.ListTasks(r.Context(), &req)
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

// GetTask 任务详情（按 id）。
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	task, err := h.logic.GetTask(r.Context(), id)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, task)
}

// GetTaskByNo 任务详情（按任务编号）。
func (h *TaskHandler) GetTaskByNo(w http.ResponseWriter, r *http.Request) {
	taskNo := r.PathValue("task_no")
	if taskNo == "" {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "task_no is required"))
		return
	}
	task, err := h.logic.GetTaskByNo(r.Context(), taskNo)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, task)
}

// ListBatchTasks 批次列表。
func (h *TaskHandler) ListBatchTasks(w http.ResponseWriter, r *http.Request) {
	var req dto.PushBatchTaskListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.ListBatchTasks(r.Context(), &req)
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

// GetBatchTask 批次详情（按 id）。
func (h *TaskHandler) GetBatchTask(w http.ResponseWriter, r *http.Request) {
	id := cast.To[uint](rest.PathParam(r, "id"))
	if id == 0 {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "invalid id"))
		return
	}
	batch, err := h.logic.GetBatchTask(r.Context(), id)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, batch)
}

// ListTasksByBatch 批次下钻子任务。
func (h *TaskHandler) ListTasksByBatch(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("batch_id")
	if batchID == "" {
		_ = rest.ErrorJSONCtx(r.Context(), w, rest.NewCodeError(http.StatusBadRequest, "batch_id is required"))
		return
	}
	var req dto.PageRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	list, total, err := h.logic.ListTasksByBatch(r.Context(), batchID, req.Page, req.PageSize)
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
