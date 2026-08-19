package middleware

import (
	"net/http"

	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/ratelimit"
)

// RateLimit Redis 分布式令牌桶限流中间件。
// 多实例部署时共享同一 Redis 键，实现全局限流。
// 若 AppAuth 已注入应用上下文，则优先使用应用配置的 RateLimit(QPS)，否则使用 fallbackRate。
func RateLimit(s *svc.ServiceContext, key string, fallbackRate, fallbackBurst float64) httpx.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rate, burst := fallbackRate, fallbackBurst
			if app, ok := AppFromContext(r.Context()); ok && app != nil && app.RateLimit > 0 {
				rate = float64(app.RateLimit)
				burst = rate * 2
			}
			limiter := ratelimit.NewRedisTokenBucket(s.Redis.Client(), "ratelimit:"+key, rate, burst)
			ok, err := limiter.AllowContext(r.Context())
			if err != nil {
				// Redis 异常时放行，避免限流器成为故障点
				next(w, r)
				return
			}
			if !ok {
				httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next(w, r)
		}
	}
}
