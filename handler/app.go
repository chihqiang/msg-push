package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// AppHandler 应用管理处理器。
type AppHandler struct {
	svc   *svc.ServiceContext
	logic *logic.AppLogic
}

// NewAppHandler 创建应用管理处理器。
func NewAppHandler(s *svc.ServiceContext) *AppHandler {
	return &AppHandler{svc: s, logic: logic.NewAppLogic(s)}
}

// Create 创建应用。
func (h *AppHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateAppRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	app, secret, err := h.logic.Create(r.Context(), &req)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, dto.AppResponse{
		ID:         app.ID,
		AppID:      app.AppID,
		Name:       app.Name,
		Secret:     secret,
		Status:     app.Status,
		DailyQuota: app.DailyQuota,
		RateLimit:  app.RateLimit,
		Remark:     app.Remark,
		CreatedAt:  app.CreatedAt,
	})
}

// List 应用列表。
func (h *AppHandler) List(w http.ResponseWriter, r *http.Request) {
	var req dto.PageRequest
	if err := httpx.BindQuery(r, &req); err != nil {
		writeBadRequest(r.Context(), w, err.Error())
		return
	}
	apps, total, err := h.logic.List(r.Context(), req.Page, req.PageSize)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, map[string]any{
		"list":      apps,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// Get 应用详情。
func (h *AppHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	app, err := h.logic.Get(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, app)
}

// Update 更新应用。
func (h *AppHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	var req dto.UpdateAppRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.Update(r.Context(), id, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}

// Delete 删除应用。
func (h *AppHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// ResetSecret 重置应用密钥。
func (h *AppHandler) ResetSecret(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}
	secret, err := h.logic.ResetSecret(r.Context(), id)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, map[string]string{"secret": secret})
}
