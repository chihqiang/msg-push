// Package e2e_test 端到端链路测试。
//
// 本文件为测试基础设施：环境启动（setupE2E）、HTTP 辅助、鉴权、通用工具，
// 以及单条/批量共用的事务链路配置（setupProviderChain）。
//
// 各链路的具体用例分散在独立文件中：
//   - single_test.go   单条消息链路
//   - batch_test.go    批量消息链路
//   - callback_test.go 短信回执回调链路
//
// 覆盖 msg-push 核心链路（走真实 HTTP + 真实 asynq Worker 消费）：
//  1. 单条消息：管理端配置 → 应用端提交 → Worker 消费 → 测试模式模拟成功 → 终态可查询
//  2. 批量消息：整批入队 → Worker 聚合处理 → 批次 completed
//  3. 短信回执回调：模拟服务商回调 → 解析落库 → 关联任务回填终态
//
// 依赖（不可用时自动 Skip）：
//   - Redis 127.0.0.1:6379（asynq 队列消费必需，可用 docker compose 启动）
//   - 数据库使用 SQLite 内存库，无需外部 MySQL
package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chihqiang/msg-push/config"
	"chihqiang/msg-push/core/pipeline"
	"chihqiang/msg-push/db"
	"chihqiang/msg-push/route"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/jwt"
	"github.com/chihqiang/infra-go/logger"
	"github.com/chihqiang/infra-go/orm"
	"github.com/chihqiang/infra-go/redisx"
	"github.com/chihqiang/infra-go/taskq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ==================== 测试环境装配 ====================

const (
	redisAddr   = "127.0.0.1:6379"
	sqliteDSN   = "file:msgpush_e2e?mode=memory&cache=shared"
	adminUser   = "admin"
	adminPass   = "admin123"
	appID       = "test-app" // 种子测试应用（is_test=true）
	appSecret   = "test-secret"
	waitTimeout = 30 * time.Second
	waitStep    = 200 * time.Millisecond
)

// e2eEnv 测试环境：HTTP baseURL + 依赖容器 + 消费端。
type e2eEnv struct {
	baseURL  string
	svc      *svc.ServiceContext
	db       *gorm.DB
	consumer *pipeline.Consumer
	ts       *httptest.Server
}

// apiResp 统一响应结构 {code, msg, data}。
type apiResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// setupE2E 装配完整环境：SQLite 内存库 + 真实 Redis + HTTP 服务 + Worker 消费端。
// 返回的环境通过 t.Cleanup 自动清理。
func setupE2E(t *testing.T) *e2eEnv {
	t.Helper()

	// 1. 探测 Redis（asynq 队列硬依赖），不可用则跳过
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		t.Skipf("redis unavailable at %s (start with docker compose up -d): %v", redisAddr, err)
	}
	_ = rdb.Close()

	// 2. 构造配置
	cfg := config.Config{
		App:    config.AppConfig{Name: "msg-push", Env: "test"},
		Server: httpx.ServerConfig{Host: "127.0.0.1", Port: 18080},
		// 日志仅输出 error，避免测试刷屏
		Logger: logger.Config{Level: logger.Level(2)},
		DB:     orm.Config{Driver: orm.DriverSQLite, DSN: sqliteDSN},
		Redis:  redisx.Config{Addr: redisAddr, KeyPrefix: "e2e:"},
		JWT:    jwt.Config{Secret: "e2e-secret", Issuer: "msg-push"},
		Taskq: taskq.Config{
			RedisAddr:       redisAddr,
			Concurrency:     10,
			DefaultMaxRetry: 3,
			DefaultTimeout:  time.Minute,
			DefaultQueue:    "default",
		},
		AccountSeed: config.AccountSeedConfig{Username: adminUser, Password: adminPass},
	}

	// 3. 初始化全局日志
	l := logger.New(cfg.Logger)
	logger.SetGlobal(l)

	// 4. 装配依赖容器
	ctx, err := svc.NewServiceContext(cfg)
	if err != nil {
		t.Fatalf("new service context: %v", err)
	}

	// 5. 建表 + 种子数据（test-app + admin 账号）
	if err := db.Migrate(ctx.DB); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Seed(ctx.DB, adminUser, adminPass); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// 6. HTTP 服务：注册路由后用 httptest 暴露（避免 httpx.Server.Start 的全局信号监听）
	server := httpx.NewServer(cfg.Server)
	route.Register(server, ctx)
	ts := httptest.NewServer(server.Handler())

	// 7. Worker 消费端：消费 msg:send / msg:batch，完成真实投递
	consumer := pipeline.NewConsumer(ctx)
	if err := consumer.Start(); err != nil {
		t.Fatalf("start consumer: %v", err)
	}

	env := &e2eEnv{baseURL: ts.URL, svc: ctx, db: ctx.DB, consumer: consumer, ts: ts}
	t.Cleanup(func() {
		consumer.Shutdown()
		ts.Close()
		ctx.Close()
		_ = l.Sync()
	})
	return env
}

// ==================== HTTP 辅助 ====================

// doJSON 发送 JSON 请求并返回 HTTP 状态码与响应体。
func (e *e2eEnv) doJSON(method, path string, body any, headers map[string]string) (int, []byte) {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, e.baseURL+path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, []byte(err.Error())
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, b
}

// parseData 解析统一响应，返回 data 原始字节。
func parseData(t *testing.T, body []byte) json.RawMessage {
	t.Helper()
	var r apiResp
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatalf("parse api resp %s: %v", body, err)
	}
	if r.Code != httpx.CodeOK {
		t.Fatalf("api error: code=%d msg=%s", r.Code, r.Msg)
	}
	return r.Data
}

// adminToken 管理端登录获取 JWT。
func (e *e2eEnv) adminToken(t *testing.T) string {
	t.Helper()
	_, body := e.doJSON("POST", "/api/v1/account/auth/login",
		map[string]string{"username": adminUser, "password": adminPass}, nil)
	var d struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(parseData(t, body), &d); err != nil {
		t.Fatalf("parse login data: %v", err)
	}
	return d.AccessToken
}

// adminHeaders 管理端鉴权请求头。
func adminHeaders(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

// appHeaders 应用端鉴权请求头（种子测试应用，明文兼容模式）。
func appHeaders() map[string]string {
	return map[string]string{"X-App-Id": appID, "X-App-Secret": appSecret}
}

// ==================== 管理端配置链路 ====================

// 通道类型与服务商（dingtalk_robot：RequiresSignature=false，无需签名映射，
// 测试模式下不真实发送，配置假地址即可走通全链路）。
const (
	e2eChannelCode = "e2e_dingtalk"
	e2eTplCode     = "e2e_tpl"
	e2eProvider    = "dingtalk_robot"
)

// idResp 解析 data.id。
type idResp struct {
	ID uint `json:"id"`
}

// setupProviderChain 走管理端 API 配置完整链路：
// 通道 → 业务模板 → 服务商账号 → 供应商模板 → 通道-模板绑定。
// 返回通道 ID（回调测试等复用）。
func (e *e2eEnv) setupProviderChain(t *testing.T) uint {
	t.Helper()
	token := e.adminToken(t)
	h := adminHeaders(token)

	// 1. 创建通道
	_, body := e.doJSON("POST", "/api/v1/account/channels", map[string]any{
		"code": e2eChannelCode, "name": "e2e 钉钉通道", "type": "dingtalk",
	}, h)
	var ch idResp
	json.Unmarshal(parseData(t, body), &ch)
	if ch.ID == 0 {
		t.Fatalf("create channel returned id=0")
	}

	// 2. 创建业务模板
	_, body = e.doJSON("POST", "/api/v1/account/templates", map[string]any{
		"code": e2eTplCode, "channel_id": ch.ID, "name": "e2e 模板",
		"content": "您好 ${name}，验证码 ${code}",
	}, h)
	var tpl idResp
	json.Unmarshal(parseData(t, body), &tpl)

	// 3. 创建服务商账号（钉钉群机器人，假 webhook；测试模式不真实调用）
	_, body = e.doJSON("POST", "/api/v1/account/provider-accounts", map[string]any{
		"account_name":  "e2e 钉钉机器人",
		"provider_code": e2eProvider,
		"config":        map[string]any{"webhook_url": "https://example.invalid/hook"},
	}, h)
	var pa idResp
	json.Unmarshal(parseData(t, body), &pa)
	if pa.ID == 0 {
		t.Fatalf("create provider account returned id=0")
	}

	// 4. 创建供应商模板
	_, body = e.doJSON("POST", "/api/v1/account/provider-templates", map[string]any{
		"provider_id":      pa.ID,
		"template_code":    "e2e_provider_tpl",
		"template_name":    "e2e 供应商模板",
		"content_type":     "text",
		"template_content": "您的验证码是 ${code}",
		"variables":        []string{"code"},
	}, h)
	var pt idResp
	json.Unmarshal(parseData(t, body), &pt)

	// 5. 创建通道-模板绑定
	_, body = e.doJSON("POST", fmt.Sprintf("/api/v1/account/channels/%d/bindings", ch.ID), map[string]any{
		"provider_template_id": pt.ID, "provider_id": pa.ID, "weight": 10, "priority": 100,
	}, h)
	var bind idResp
	json.Unmarshal(parseData(t, body), &bind)
	if bind.ID == 0 {
		t.Fatalf("create binding returned id=0")
	}

	return ch.ID
}

// ==================== 轮询辅助 ====================

// waitFor 轮询直到条件满足或超时。
func waitFor(t *testing.T, desc string, cond func() (bool, string)) {
	t.Helper()
	deadline := time.Now().Add(waitTimeout)
	for time.Now().Before(deadline) {
		if ok, _ := cond(); ok {
			return
		}
		time.Sleep(waitStep)
	}
	t.Fatalf("timeout waiting for %s", desc)
}
