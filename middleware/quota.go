package middleware

import (
	"net/http"

	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// Quota 配额中间件：按应用每日配额校验并原子计数。
// 依赖 AppAuth 已注入应用上下文；DailyQuota<=0 表示不限制；配额服务异常时放行避免故障点。
func Quota(s *svc.ServiceContext) httpx.Middleware {
	q := logic.NewQuotaLogic(s)
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			app, ok := AppFromContext(r.Context())
			if !ok || app == nil {
				next(w, r)
				return
			}
			if app.DailyQuota <= 0 {
				next(w, r)
				return
			}
			allowed, err := q.Check(r.Context(), app.ID, app.DailyQuota)
			if err != nil {
				// 配额服务异常放行，避免配额器成为故障点
				next(w, r)
				return
			}
			if !allowed {
				httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusTooManyRequests, "daily quota exceeded")
				return
			}
			next(w, r)
		}
	}
}
