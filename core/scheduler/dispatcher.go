// 本文件实现 Webhook 异步投递器：轮询 outbox 日志，认领并投递，失败退避重试。
package scheduler

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
)

// Dispatcher 配置。
const (
	dispatcherInterval       = time.Second      // 轮询间隔
	dispatcherBatch          = 100              // 每批处理数
	dispatcherLeaseBase      = 10 * time.Second // 认领租约基础时长
	dispatcherMaxConcurrency = 50               // 最大并发投递数（信号量上限）
)

// dispatcherHTTPClient Webhook 投递共用 HTTP 客户端（复用连接池）。
// 无固定 Timeout：整体超时由 send 依据 WebhookConfig.Timeout 设置的 ctx 控制。
var dispatcherHTTPClient = &http.Client{}

// Dispatcher Webhook 异步投递器（outbox 模式）。
// 轮询 pending/processing 到期日志 → CAS 认领 → 并发投递 → 成功置终态，失败退避重试。
type Dispatcher struct {
	svc    *svc.ServiceContext
	sem    chan struct{}  // 并发投递信号量，防止 webhook 慢时 goroutine 无限堆积
	wg     sync.WaitGroup // 跟踪在途投递协程，优雅停止时等待其完成
	stopCh chan struct{}
	doneCh chan struct{}
}

// NewDispatcher 创建投递器。
func NewDispatcher(s *svc.ServiceContext) *Dispatcher {
	return &Dispatcher{
		svc:    s,
		sem:    make(chan struct{}, dispatcherMaxConcurrency),
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// Start 启动投递器（非阻塞，后台轮询）。
func (d *Dispatcher) Start() {
	go d.run()
}

// Stop 优雅停止。等待调度循环退出后再等待在途投递完成，
// 避免强制退出导致已认领（processing + 租约未到期）的日志悬挂。
func (d *Dispatcher) Stop() {
	close(d.stopCh)
	<-d.doneCh
	d.wg.Wait()
}

// StartService 适配 service.Starter。
func (d *Dispatcher) StartService() { d.Start() }

// run 主循环。
func (d *Dispatcher) run() {
	defer close(d.doneCh)
	ticker := time.NewTicker(dispatcherInterval)
	defer ticker.Stop()
	// 启动立即投递一次
	d.dispatchOnce()
	for {
		select {
		case <-ticker.C:
			d.dispatchOnce()
		case <-d.stopCh:
			return
		}
	}
}

// dispatchOnce 处理一批到期日志。
func (d *Dispatcher) dispatchOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	var due []model.WebhookLog
	err := d.svc.DB.WithContext(ctx).
		Where("(status = ? AND next_attempt_at <= ?) OR (status = ? AND locked_until <= ?)",
			string(model.WebhookLogPending), now,
			string(model.WebhookLogProcessing), now).
		Order("id ASC").
		Limit(dispatcherBatch).
		Find(&due).Error
	if err != nil {
		logger.Warnf("webhook dispatcher: query due failed: %v", err)
		return
	}

	for i := range due {
		logEntry := &due[i]
		// CAS 认领：租约基于该日志配置的超时时间，确保租约覆盖投递耗时，
		// 避免 Timeout > 租约时长时被下轮重复认领导致同一 webhook 并发重复投递。
		claimed, token := d.claim(ctx, logEntry, leaseDuration(logEntry.TimeoutSeconds))
		if !claimed {
			continue
		}
		// 并发限流：信号量满则跳过本批剩余（已认领日志会在租约到期后下轮重新扫描）
		select {
		case d.sem <- struct{}{}:
		default:
			continue
		}
		d.wg.Add(1)
		go func() {
			defer func() { <-d.sem }()
			defer d.wg.Done()
			d.send(ctx, logEntry, token)
		}()
	}
}

// claim CAS 认领日志：仅当状态匹配且租约未到期时置 processing。
func (d *Dispatcher) claim(ctx context.Context, logEntry *model.WebhookLog, lease time.Duration) (bool, string) {
	token := randomToken()
	lockedUntil := time.Now().Add(lease)
	now := time.Now()

	// 认领条件：pending 且未到期；或 processing 且租约过期
	res := d.svc.DB.WithContext(ctx).Model(&model.WebhookLog{}).
		Where("id = ? AND ((status = ? AND next_attempt_at <= ?) OR (status = ? AND locked_until <= ?))",
			logEntry.ID,
			string(model.WebhookLogPending), now,
			string(model.WebhookLogProcessing), now).
		Updates(map[string]any{
			"status":       string(model.WebhookLogProcessing),
			"lease_token":  token,
			"locked_until": lockedUntil,
			"updated_at":   time.Now(),
		})
	if res.Error != nil || res.RowsAffected == 0 {
		return false, ""
	}
	return true, token
}

// send 投递 webhook（带认领令牌的条件更新）。
func (d *Dispatcher) send(ctx context.Context, logEntry *model.WebhookLog, token string) {
	// 独立超时，避免拖垮调度循环
	timeout := time.Duration(logEntry.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	sendCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	attemptTime := time.Now()
	statusCode, respBody, err := d.doPost(sendCtx, logEntry, attemptTime)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()

	if err == nil && statusCode >= 200 && statusCode < 300 {
		// 成功
		_ = d.svc.DB.WithContext(dbCtx).Model(&model.WebhookLog{}).
			Where("id = ? AND lease_token = ?", logEntry.ID, token).
			Updates(map[string]any{
				"status":          string(model.WebhookLogSuccess),
				"response_status": statusCode,
				"response_data":   truncate(string(respBody), 1<<20),
				"error_message":   "",
				"locked_until":    nil,
				"lease_token":     "",
				"updated_at":      time.Now(),
			}).Error
		return
	}

	// 失败：处理错误信息
	errMsg := "http status " + strconv.Itoa(statusCode)
	if err != nil {
		errMsg = err.Error()
	}
	if statusCode == 0 {
		statusCode = 0
	}

	// 失败：若重试未超限则退避重试，否则置失败
	if logEntry.RetryCount < logEntry.MaxRetries {
		nextAttempt := time.Now().Add(backoffDelay(logEntry.RetryCount))
		_ = d.svc.DB.WithContext(dbCtx).Model(&model.WebhookLog{}).
			Where("id = ? AND lease_token = ?", logEntry.ID, token).
			Updates(map[string]any{
				"status":          string(model.WebhookLogPending),
				"retry_count":     logEntry.RetryCount + 1,
				"response_status": statusCode,
				"response_data":   truncate(string(respBody), 1<<20),
				"error_message":   truncate(errMsg, 1024),
				"next_attempt_at": nextAttempt,
				"locked_until":    nil,
				"lease_token":     "",
				"updated_at":      time.Now(),
			}).Error
		logger.Warnf("webhook dispatcher: send to %s failed (retry %d/%d): %s", logEntry.WebhookURL, logEntry.RetryCount+1, logEntry.MaxRetries, errMsg)
		return
	}

	// 重试耗尽：置失败终态
	_ = d.svc.DB.WithContext(dbCtx).Model(&model.WebhookLog{}).
		Where("id = ? AND lease_token = ?", logEntry.ID, token).
		Updates(map[string]any{
			"status":          string(model.WebhookLogFailed),
			"response_status": statusCode,
			"response_data":   truncate(string(respBody), 1<<20),
			"error_message":   truncate(errMsg, 1024),
			"locked_until":    nil,
			"lease_token":     "",
			"updated_at":      time.Now(),
		}).Error
	logger.Warnf("webhook dispatcher: send to %s failed permanently: %s", logEntry.WebhookURL, errMsg)
}

// doPost 执行 HTTP POST 并携带签名头。
// 客户端超时受 ctx 控制（由 send 按 WebhookConfig.Timeout 设置），故用超时上下文而非
// 固定 client.Timeout，使 Webhook 配置的超时值真正生效。
func (d *Dispatcher) doPost(ctx context.Context, logEntry *model.WebhookLog, attemptTime time.Time) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logEntry.WebhookURL, bytes.NewReader([]byte(logEntry.RequestData)))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MessagePush-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", logEntry.Event)
	req.Header.Set("X-Webhook-Timestamp", strconv.FormatInt(attemptTime.Unix(), 10))
	req.Header.Set("X-Webhook-Delivery-ID", strconv.FormatUint(uint64(logEntry.ID), 10))
	if logEntry.SigningSecret != "" {
		req.Header.Set("X-Webhook-Signature", webhookSignature([]byte(logEntry.RequestData), logEntry.SigningSecret, attemptTime.Unix()))
	}

	// 复用全局连接池客户端；整体超时由 ctx（含 WebhookConfig.Timeout）控制
	resp, err := dispatcherHTTPClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, body, nil
}

// webhookSignature 生成 HMAC-SHA256 签名：hex(HMAC_SHA256(secret, "<timestamp>.<body>"))。
func webhookSignature(body []byte, secret string, timestamp int64) string {
	msg := strconv.FormatInt(timestamp, 10) + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

// backoffDelay 退避延迟：线性增长（第 n 次重试延迟 n 秒）。
func backoffDelay(retryCount int) time.Duration {
	seconds := retryCount + 1
	if seconds > 60 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}

// leaseDuration 认领租约时长（基于超时时间）。
func leaseDuration(timeoutSeconds int) time.Duration {
	lease := dispatcherLeaseBase + time.Duration(timeoutSeconds)*time.Second
	if timeoutSeconds <= 0 {
		lease = 30 * time.Second
	}
	return lease
}

// randomToken 生成随机认领令牌。
func randomToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// truncate 截断字节（不切断 UTF-8 多字节字符）。
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return string([]byte(s)[:max])
}
