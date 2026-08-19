// Package middleware 提供应用鉴权、管理端鉴权、限流等 HTTP 中间件。
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/hash"
	"github.com/chihqiang/infra-go/httpx"
	"gorm.io/gorm"
)

// ctxKey 上下文键类型，避免与第三方库 key 冲突。
type ctxKey string

// contextKeyApp 当前请求应用信息的上下文键。
const contextKeyApp ctxKey = "msgpush.app"

// hmacAllowedSkew 时间戳防重放允许偏差（秒）：±5 分钟。
const hmacAllowedSkew = 300

// nonceTTL nonce 防重放缓存有效期（秒）。
const nonceTTL = 300

// AppFromContext 从请求上下文获取当前应用。
func AppFromContext(ctx context.Context) (*model.Application, bool) {
	app, ok := ctx.Value(contextKeyApp).(*model.Application)
	return app, ok
}

// AppAuth 应用鉴权中间件：校验应用身份。
// 支持两种认证方式（二选一）：
//  1. HMAC-SHA256 签名认证：请求携带 X-Signature / X-Timestamp / X-Nonce，
//     签名 = HMAC-SHA256(secret, method+path+sortedParams+timestamp+nonce)，含时间戳与 nonce 防重放；
//  2. 兼容模式：请求携带 X-App-Secret 明文，bcrypt 比对。
//
// 校验通过后将应用注入请求上下文，供 handler/logic 使用。
func AppAuth(s *svc.ServiceContext) httpx.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			appID := r.Header.Get("X-App-Id")
			if appID == "" {
				httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, "missing app credentials")
				return
			}

			var app model.Application
			err := s.DB.WithContext(r.Context()).
				Where("app_id = ? AND status = 1", appID).
				First(&app).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, "invalid app credentials")
				return
			}
			if err != nil {
				httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusInternalServerError, "internal error")
				return
			}

			// 认证：优先 HMAC 签名，否则兼容 X-App-Secret 简单校验
			signature := r.Header.Get("X-Signature")
			timestampStr := r.Header.Get("X-Timestamp")
			nonce := r.Header.Get("X-Nonce")
			if signature != "" || timestampStr != "" || nonce != "" {
				if !verifyHMACRequest(s, &app, r, signature, timestampStr, nonce) {
					httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, "invalid signature")
					return
				}
			} else {
				appSecret := r.Header.Get("X-App-Secret")
				if appSecret == "" || !hash.BcryptMatch(app.AppSecret, appSecret) {
					httpx.WriteHTTPErrorCtx(r.Context(), w, http.StatusUnauthorized, "invalid app credentials")
					return
				}
			}

			ctx := context.WithValue(r.Context(), contextKeyApp, &app)
			next(w, r.WithContext(ctx))
		}
	}
}

// verifyHMACRequest 校验 HMAC 签名请求：时间戳 ±5 分钟、nonce 唯一、签名匹配。
func verifyHMACRequest(s *svc.ServiceContext, app *model.Application, r *http.Request, signature, timestampStr, nonce string) bool {
	if app.AppSecretPlain == "" {
		return false
	}
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}

	// 时间戳防重放：允许 ±5 分钟
	now := time.Now().Unix()
	diff := now - timestamp
	if diff > hmacAllowedSkew || diff < -hmacAllowedSkew {
		return false
	}

	// nonce 防重放：Redis SETNX，同一 app+nonce 仅首次放行
	nonceKey := fmt.Sprintf("msgpush:nonce:%s:%s", app.AppID, nonce)
	ok, err := s.Redis.Client().SetNX(r.Context(), nonceKey, 1, nonceTTL*time.Second).Result()
	if err != nil || !ok {
		return false
	}

	// 读取请求体用于签名验证，读完放回以便后续 handler 读取
	var bodyParams map[string]any
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			return false
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &bodyParams); err != nil {
				return false
			}
		}
	}

	return verifySignature(app.AppSecretPlain, r.Method, r.URL.Path, bodyParams, timestamp, nonce, signature)
}
