// msg-push 统一消息推送服务
//
// 单进程内同时提供：
//   - HTTP API：应用鉴权（HMAC/密钥）、消息发送/批量发送/任务查询、管理端（应用/通道/模板/服务商管理）、统计分析
//   - 消费端 Worker：消费 taskq(asynq/Redis) 入队的投递任务，完成选通道、模板渲染、多服务商发送、重试/切供应商
//   - 后台调度：Webhook 异步投递、短信回执超时扫描、短信状态主动查询、配额统计同步
//
// 依赖组件经 svc.ServiceContext 统一装配，全部后台服务通过 servicex 适配器纳入
// service.NewServiceGroup 统一启停，支持优雅关闭。
package main

import (
	"flag"
	"os"
	"path/filepath"

	"chihqiang/msg-push/config"
	"chihqiang/msg-push/db"
	"chihqiang/msg-push/route"
	"chihqiang/msg-push/servicex"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/conf"
	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
)

func main() {
	// 1. 加载配置
	var cfg config.Config
	configFile := flag.String("config", "config.yaml", "config file path")
	flag.Parse()
	conf.MustLoad(*configFile, &cfg, conf.UseEnv())

	// 2. 初始化日志（全局）
	l := logger.New(cfg.Logger)
	logger.SetGlobal(l)
	defer func() { _ = l.Sync() }()

	// SQLite 文件数据库需要确保目录存在（在打开数据库之前）
	if cfg.DB.Driver == "sqlite" && cfg.DB.Database != "" && cfg.DB.Database != ":memory:" {
		_ = os.MkdirAll(filepath.Dir(cfg.DB.Database), 0o755)
	}

	// 3. 装配依赖
	ctx, err := svc.NewServiceContext(cfg)
	if err != nil {
		logger.Fatalf("init service context: %v", err)
	}
	defer ctx.Close()

	// 4. 迁移 + 种子数据
	if err := db.Migrate(ctx.DB); err != nil {
		logger.Fatalf("migrate db: %v", err)
	}
	if err := db.Seed(ctx.DB, cfg.AccountSeed.Username, cfg.AccountSeed.Password); err != nil {
		logger.Fatalf("seed db: %v", err)
	}

	// 5. HTTP 服务 + 路由
	server := httpx.NewServer(cfg.Server)
	route.Register(server, ctx)

	// 6. 后台服务统一装配（消费端 / Webhook 投递 / 定时调度 / HTTP API），统一启停、优雅关闭
	sg := servicex.NewServiceGroup(ctx, server)
	sg.Start()
}
