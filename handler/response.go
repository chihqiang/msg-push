// Package handler 处理器层：参数绑定 → 调用 Logic → 统一响应（薄层）。
package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/jwt"
)

// writeError 统一业务错误响应（code=-1, msg=err）。
func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	httpx.OkJSONCtx(ctx, w, httpx.NewCodeError(httpx.CodeDefaultError, err.Error()))
}

// writeBadRequest 参数/资源错误响应（code=400）。
func writeBadRequest(ctx context.Context, w http.ResponseWriter, msg string) {
	httpx.OkJSONCtx(ctx, w, httpx.NewCodeError(httpx.CodeBadRequest, msg))
}

// accountIDFromCtx 从 JWT claims 获取当前商户 ID。
func accountIDFromCtx(r *http.Request) (uint, bool) {
	claims := jwt.ClaimsFromContext(r.Context())
	if claims == nil {
		return 0, false
	}
	raw, _ := claims[jwt.ClaimKeyUserID].(string)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}
