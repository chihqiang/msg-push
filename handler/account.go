package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// AccountHandler 商户认证处理器。
type AccountHandler struct {
	svc   *svc.ServiceContext
	logic *logic.AccountLogic
}

// NewAccountHandler 创建商户认证处理器。
func NewAccountHandler(s *svc.ServiceContext) *AccountHandler {
	return &AccountHandler{svc: s, logic: logic.NewAccountLogic(s)}
}

// Login 商户登录。
func (h *AccountHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.AccountLoginRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	resp, err := h.logic.Login(r.Context(), &req)
	if err != nil {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, err.Error())
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// Refresh 刷新令牌。
func (h *AccountHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.AccountRefreshRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	resp, err := h.logic.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// Me 获取当前登录商户个人资料。
func (h *AccountHandler) Me(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountIDFromCtx(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, "missing account context")
		return
	}
	resp, err := h.logic.Profile(r.Context(), accountID)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, resp)
}

// ChangePassword 修改当前登录商户密码。
func (h *AccountHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	accountID, ok := accountIDFromCtx(r)
	if !ok {
		httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, "missing account context")
		return
	}
	var req dto.ChangePasswordRequest
	if err := httpx.MustBindJSON(w, r, &req); err != nil {
		return
	}
	if err := h.logic.ChangePassword(r.Context(), accountID, &req); err != nil {
		writeError(r.Context(), w, err)
		return
	}
	httpx.OkJSONCtx(r.Context(), w, nil)
}
