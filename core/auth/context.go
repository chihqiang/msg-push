package auth

import "context"

type claimsContextKey struct{}

// WithClaims 把校验通过的令牌载荷写入请求上下文。
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, claimsContextKey{}, claims)
}

// ClaimsFrom 读取 WithClaims 写入的令牌载荷。
func ClaimsFrom(ctx context.Context) (*Claims, bool) {
	if ctx == nil {
		return nil, false
	}
	claims, ok := ctx.Value(claimsContextKey{}).(*Claims)
	return claims, ok && claims != nil
}
