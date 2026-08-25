// Package servicex 提供各后台服务的 service.Service 适配器，并统一装配进
// service.NewServiceGroup 编排，支持优雅关闭。
//
// main 只需传入已装配依赖的 *svc.ServiceContext 与已注册路由的 *httpx.Server，
// 通过 NewServiceGroup 即可创建全部后台服务（消费端 / Webhook 投递 / 定时调度 /
// HTTP API）并统一启停，无需逐个初始化。
package servicex

import (
	"chihqiang/msg-push/scheduler"
	"chihqiang/msg-push/svc"
	"chihqiang/msg-push/webhookx"
	"chihqiang/msg-push/worker"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
	"github.com/chihqiang/infra-go/service"
)

// NewServiceGroup 装配全部后台服务并返回统一的 service.ServiceGroup。
//
// 内部完成以下服务的创建、适配与注册：
//   - worker.Consumer：消费 msg:send / msg:batch，完成实际投递
//   - webhookx.Dispatcher：Webhook 异步投递器（outbox 模式，含重试退避）
//   - scheduler.SMSTimeoutScanner：短信回执超时扫描器
//   - scheduler.QuotaSyncer：配额统计同步器
//   - scheduler.StatusPuller：短信状态主动查询扫描器（服务商无回执时主动补单）
//   - HTTP API 服务（由 main 创建并注册路由后传入）
//
// 返回的 *service.ServiceGroup 通过 Start() 统一启动，Stop() 优雅关闭。
func NewServiceGroup(ctx *svc.ServiceContext, server *httpx.Server) *service.ServiceGroup {
	sg := service.NewServiceGroup()

	sg.Add(NewQuotaSyncerService(scheduler.NewQuotaSyncer(ctx)))
	sg.Add(NewSMSTimeoutScannerService(scheduler.NewSMSTimeoutScanner(ctx)))
	sg.Add(NewStatusPullerService(scheduler.NewStatusPuller(ctx)))
	sg.Add(NewWebhookDispatcherService(webhookx.NewDispatcher(ctx)))
	sg.Add(NewConsumerService(worker.NewConsumer(ctx)))
	sg.Add(service.WithStart(func() {
		logger.Infof("msg-push api listening on %s:%d (env=%s)",
			ctx.Config.Server.Host, ctx.Config.Server.Port, ctx.Config.App.Env)
		if err := server.Start(); err != nil {
			logger.Errorf("http server stopped: %v", err)
		}
	}))

	return sg
}
