package common

import (
	"context"
	"strconv"

	"chihqiang/msg-push/core/auth"
)

// AccountIDFromContext 从 JWT claims 获取当前商户 ID。
func AccountIDFromContext(ctx context.Context) (uint, bool) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return 0, false
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}
