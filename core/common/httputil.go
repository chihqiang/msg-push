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

// ParseID 解析路径参数中的 uint ID。
func ParseID(r *http.Request) (uint, error) {
	return ParseIDValue(r, "id")
}

// ParseIDValue 解析指定名称的路径参数为 uint ID。
// 用于嵌套路由（如 /channels/{id}/bindings/{binding_id}）取非首段参数。
func ParseIDValue(r *http.Request, name string) (uint, error) {
	raw := r.PathValue(name)
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

// ChannelIDFromPath 解析路径中的通道 ID。
func ChannelIDFromPath(r *http.Request) (uint, bool) {
	channelID, err := ParseID(r)
	if err != nil {
		return 0, false
	}
	return channelID, true
}
