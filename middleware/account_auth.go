package middleware

import (
	"net/http"
	"strings"

	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// AccountAuth 商户 JWT 鉴权中间件。
// 从 Authorization: Bearer <token> 提取令牌，验证后由 jwt 包将 claims 注入上下文。
func AccountAuth(s *svc.ServiceContext) httpx.Middleware {
	return s.JWT.AuthMiddleware(func(r *http.Request) string {
		auth := r.Header.Get("Authorization")
		return strings.TrimPrefix(auth, "Bearer ")
	})
}
