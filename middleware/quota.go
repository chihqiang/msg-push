package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// Quota 配额中间件：按应用每日配额校验并原子计数。
// 依赖 AppAuth 已注入应用上下文；DailyQuota<=0 表示不限制；配额服务异常时放行避免故障点。
// 仅发送类请求（POST /messages、/messages/batch）消耗配额，查询类（GET）不消耗；
// 批量请求按接收者数量一次性扣减，避免一条批量请求绕过每日配额。
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
			// 仅发送类请求消耗配额，查询任务不消耗
			if r.Method != http.MethodPost {
				next(w, r)
				return
			}
			count := 1
			if strings.HasSuffix(r.URL.Path, "/messages/batch") {
				if n := batchReceiverCount(r); n > 0 {
					count = n
				}
			}
			allowed, err := q.CheckN(r.Context(), app.ID, app.DailyQuota, count)
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

// batchReceiverCount 从批量请求体解析接收者数量（尽力解析，失败返回 0）。
// AppAuth 可能已读取过 body 用于签名校验并放回，这里再读后同样放回，保证后续 handler 可读。
func batchReceiverCount(r *http.Request) int {
	if r.Body == nil {
		return 0
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return 0
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	var req struct {
		Receivers []string `json:"receivers"`
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return 0
	}
	return len(req.Receivers)
}
