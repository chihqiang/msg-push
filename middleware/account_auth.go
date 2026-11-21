package middleware

import (
	"errors"
	"net/http"

	"github.com/chihqiang/ten-go/rest/chain"
	resthandler "github.com/chihqiang/ten-go/rest/handler"

	"chihqiang/msg-push/core/auth"
	"chihqiang/msg-push/svc"
)

// AccountAuth 商户 JWT 鉴权中间件。
// 从 Authorization: Bearer <token> 提取访问令牌，校验通过后将 claims 注入请求上下文。
func AccountAuth(s *svc.ServiceContext) chain.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			raw, err := resthandler.BearerToken(r)
			if err != nil {
				reject(w, err)
				return
			}
			claims, err := s.Tokens.ParseAccess(raw)
			if err != nil {
				reject(w, err)
				return
			}
			next(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		}
	}
}

// reject 按 RFC 6750 返回无响应体的鉴权失败。
func reject(w http.ResponseWriter, err error) {
	status := http.StatusUnauthorized
	if errors.Is(err, resthandler.ErrMalformedAuthorization) {
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)
}
