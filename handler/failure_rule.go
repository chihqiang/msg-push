package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// FailureRuleHandler 失败规则管理处理器。
type FailureRuleHandler struct {
	svc   *svc.ServiceContext
	logic *logic.FailureRuleLogic
}

// NewFailureRuleHandler 创建失败规则管理处理器。
func NewFailureRuleHandler(s *svc.ServiceContext) *FailureRuleHandler {
	return &FailureRuleHandler{svc: s, logic: logic.NewFailureRuleLogic(s)}
}

// Create 创建失败规则。
func (h *FailureRuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateFailureRuleRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	rule, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, rule)
}

// List 失败规则列表。
func (h *FailureRuleHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.FailureRuleListRequest
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

// Get 失败规则详情。
func (h *FailureRuleHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	rule, err := h.logic.Get(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, rule)
}

// Update 更新失败规则。
func (h *FailureRuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	var req dto.UpdateFailureRuleRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除失败规则。
func (h *FailureRuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	if err := h.logic.Delete(r.Context(), id); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// GetOptions 获取失败规则选项。
func (h *FailureRuleHandler) GetOptions(w http.ResponseWriter, r *http.Request) {
	httpx.OkJSONCtx(r.Context(), w, h.logic.GetOptions(r.Context()))
}

// RefreshCache 刷新规则缓存。
func (h *FailureRuleHandler) RefreshCache(w http.ResponseWriter, r *http.Request) {
	if err := h.logic.RefreshCache(r.Context()); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}
