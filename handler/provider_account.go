package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// ProviderAccountHandler 服务商账号管理处理器。
type ProviderAccountHandler struct {
	svc   *svc.ServiceContext
	logic *logic.ProviderAccountLogic
}

// NewProviderAccountHandler 创建服务商账号管理处理器。
func NewProviderAccountHandler(s *svc.ServiceContext) *ProviderAccountHandler {
	return &ProviderAccountHandler{svc: s, logic: logic.NewProviderAccountLogic(s)}
}

// GetAvailableProviders 获取可用服务商列表。
func (h *ProviderAccountHandler) GetAvailableProviders(w http.ResponseWriter, r *http.Request) {
	var req dto.AvailableProvidersRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	httpx.OkJSONCtx(r.Context(), w, h.logic.GetAvailableProviders(req.ProviderType))
}

// GetProviderConfigFields 获取服务商配置字段定义。
func (h *ProviderAccountHandler) GetProviderConfigFields(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("provider_code")
	fields, err := h.logic.GetProviderConfigFields(code)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, fields)
}

// Create 创建服务商账号。
func (h *ProviderAccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProviderAccountRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	resp, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// List 服务商账号列表。
func (h *ProviderAccountHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.ProviderAccountListRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	resp, err := h.logic.List(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// Get 服务商账号详情。
func (h *ProviderAccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	resp, err := h.logic.Get(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// Update 更新服务商账号。
func (h *ProviderAccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	var req dto.UpdateProviderAccountRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除服务商账号。
func (h *ProviderAccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
