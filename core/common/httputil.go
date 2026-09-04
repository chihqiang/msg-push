package common

import (
	"context"
	"net/http"
	"strconv"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/jwt"
)

// WriteError 统一业务错误响应（code=-1, msg=err）。
func WriteError(ctx context.Context, w http.ResponseWriter, err error) {
	httpx.OkJSONCtx(ctx, w, httpx.NewCodeError(httpx.CodeDefaultError, err.Error()))
}

// WriteBadRequest 参数/资源错误响应（code=400）。
func WriteBadRequest(ctx context.Context, w http.ResponseWriter, msg string) {
	httpx.OkJSONCtx(ctx, w, httpx.NewCodeError(httpx.CodeBadRequest, msg))
}

// AccountIDFromContext 从 JWT claims 获取当前商户 ID。
func AccountIDFromContext(ctx context.Context) (uint, bool) {
	claims := jwt.ClaimsFromContext(ctx)
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
