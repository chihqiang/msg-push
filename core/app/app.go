// Package app 提供后台服务装配：将各核心组件（消费端 / 定时调度 / Webhook 投递 / HTTP API）
// 适配为 service.Service 并纳入统一的 service.ServiceGroup，支持优雅关闭。
//
// main 只需传入已装配依赖的 *svc.ServiceContext 与已注册路由的 *httpx.Server，
// 调用 NewServiceGroup 即可完成全部后台服务的统一启停。
package app

import (
	"chihqiang/msg-push/core/pipeline"
	"chihqiang/msg-push/core/scheduler"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/logger"
	"github.com/chihqiang/infra-go/service"
)

// serviceAdapter 将非标准接口的后台组件适配为 service.Service。
// 目前仅 pipeline.Consumer 的 Start 返回 error，不满足 service.Starter；
// 其余组件（QuotaSyncer/SMSTimeoutScanner/StatusPuller/Dispatcher）均已原生实现无返回值的 Start()/Stop()。
type serviceAdapter struct {
	start func()
	stop  func()
}

func (a serviceAdapter) Start() { a.start() }
func (a serviceAdapter) Stop()  { a.stop() }

// NewServiceGroup 装配全部后台服务并返回统一的 service.ServiceGroup。
//
// 内部完成以下服务的创建与注册：
//   - pipeline.Consumer：消费 msg:send / msg:batch，完成实际投递（经 serviceAdapter 适配）
//   - scheduler.Dispatcher：Webhook 异步投递器（outbox 模式，含重试退避）
//   - scheduler.SMSTimeoutScanner：短信回执超时扫描器
//   - scheduler.QuotaSyncer：配额统计同步器
//   - scheduler.StatusPuller：短信状态主动查询扫描器（服务商无回执时主动补单）
//   - HTTP API 服务（由 main 创建并注册路由后传入）
//
// 返回的 *service.ServiceGroup 通过 Start() 统一启动，Stop() 优雅关闭。
func NewServiceGroup(ctx *svc.ServiceContext, server *httpx.Server) *service.ServiceGroup {
	sg := service.NewServiceGroup()

	sg.Add(scheduler.NewQuotaSyncer(ctx))
	sg.Add(scheduler.NewSMSTimeoutScanner(ctx))
	sg.Add(scheduler.NewStatusPuller(ctx))
	sg.Add(scheduler.NewDispatcher(ctx))
	consumer := pipeline.NewConsumer(ctx)
	sg.Add(serviceAdapter{start: consumer.StartService, stop: consumer.Stop})
	sg.Add(service.WithStart(func() {
		logger.Infof("msg-push api listening on %s:%d (env=%s)",
			ctx.Config.Server.Host, ctx.Config.Server.Port, ctx.Config.App.Env)
		if err := server.Start(); err != nil {
			logger.Errorf("http server stopped: %v", err)
		}
	}))

	return sg
}
