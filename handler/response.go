// Package handler 处理器层：参数绑定 → 调用 Logic → 统一响应（薄层）。
package handler

import (
	"context"
	"net/http"

	"chihqiang/msg-push/core/common"
)

// writeError 统一业务错误响应（code=-1, msg=err）。
func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	common.WriteError(ctx, w, err)
}

// writeBadRequest 参数/资源错误响应（code=400）。
func writeBadRequest(ctx context.Context, w http.ResponseWriter, msg string) {
	common.WriteBadRequest(ctx, w, msg)
}

// accountIDFromCtx 从 JWT claims 获取当前商户 ID。
func accountIDFromCtx(r *http.Request) (uint, bool) {
	return common.AccountIDFromContext(r.Context())
}
