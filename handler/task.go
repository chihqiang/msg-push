package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
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
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.ListTasks(r.Context(), &req)
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

// GetTask 任务详情（按 id）。
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := httpx.PathValue(r, "id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	task, err := h.logic.GetTask(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, task)
}

// GetTaskByNo 任务详情（按任务编号）。
func (h *TaskHandler) GetTaskByNo(w http.ResponseWriter, r *http.Request) {
	taskNo := r.PathValue("task_no")
	if taskNo == "" {
		writeBadRequest(r.Context(), w, "task_no is required")
		return
	}
	task, err := h.logic.GetTaskByNo(r.Context(), taskNo)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, task)
}

// ListBatchTasks 批次列表。
func (h *TaskHandler) ListBatchTasks(w http.ResponseWriter, r *http.Request) {
	var req dto.PushBatchTaskListRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.ListBatchTasks(r.Context(), &req)
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

// GetBatchTask 批次详情（按 id）。
func (h *TaskHandler) GetBatchTask(w http.ResponseWriter, r *http.Request) {
	id := httpx.PathValue(r, "id", uint(0))
	if id == 0 {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	batch, err := h.logic.GetBatchTask(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, batch)
}

// ListTasksByBatch 批次下钻子任务。
func (h *TaskHandler) ListTasksByBatch(w http.ResponseWriter, r *http.Request) {
	batchID := r.PathValue("batch_id")
	if batchID == "" {
		writeBadRequest(r.Context(), w, "batch_id is required")
		return
	}
	var req dto.PageRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	list, total, err := h.logic.ListTasksByBatch(r.Context(), batchID, req.Page, req.PageSize)
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
