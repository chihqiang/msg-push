// Package route 统一注册路由与中间件链。
package route

import (
	"net/http"

	"chihqiang/msg-push/handler"
	"chihqiang/msg-push/middleware"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// Register 注册全部路由与全局/分组中间件。
func Register(server *httpx.Server, s *svc.ServiceContext) {
	// 全局中间件
	server.Use(httpx.WithRecovery())
	server.Use(httpx.WithRequestID())
	server.Use(httpx.WithLogger())
	server.Use(httpx.WithCors("*"))

	// 健康检查
	server.AddRoute(httpx.Route{
		Method: http.MethodGet,
		Path:   "/health",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			httpx.OkJSON(w, map[string]string{"status": "ok"})
		},
	})

	mh := handler.NewMessageHandler(s)
	ah := handler.NewAccountHandler(s)
	appH := handler.NewAppHandler(s)
	quotaH := handler.NewQuotaHandler(s)
	chH := handler.NewChannelHandler(s)
	tplH := handler.NewTemplateHandler(s)
	paH := handler.NewProviderAccountHandler(s)
	psH := handler.NewProviderSignatureHandler(s)
	ptH := handler.NewProviderTemplateHandler(s)
	cbH := handler.NewChannelBindingHandler(s)
	csmH := handler.NewChannelSignatureMappingHandler(s)
	ctH := handler.NewChannelTestHandler(s)
	frH := handler.NewFailureRuleHandler(s)
	callbackH := handler.NewCallbackHandler(s)
	whH := handler.NewWebhookHandler(s)
	logH := handler.NewLogHandler(s)
	taskH := handler.NewTaskHandler(s)
	statH := handler.NewStatisticHandler(s)

	// ============ 服务商回调（公开路由，供服务商调用，无需认证） ============
	server.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/api/callback/{id}", Handler: callbackH.Handle})
	server.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/api/callback/{id}", Handler: callbackH.Handle})

	// ============ 应用侧 API：消息发送（应用鉴权 + 配额 + 限流） ============
	api := server.Group("/api/v1",
		middleware.AppAuth(s),
		middleware.Quota(s),
		middleware.RateLimit(s, "message", 100, 200),
	)
	api.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/messages", Handler: mh.Send})
	api.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/messages/batch", Handler: mh.BatchSend})
	api.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/tasks/{task_id}", Handler: mh.QueryTask})

	// ============ 账号 API：认证（公开） ============
	account := server.Group("/api/v1/account")
	account.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/auth/login", Handler: ah.Login})
	account.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/auth/refresh", Handler: ah.Refresh})

	// ============ 账号 API：受保护（JWT 鉴权） ============
	protected := server.Group("/api/v1/account", middleware.AccountAuth(s))

	// 个人中心
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/me", Handler: ah.Me})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/password", Handler: ah.ChangePassword})

	// 应用管理
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/apps", Handler: appH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/apps", Handler: appH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/apps/{id}", Handler: appH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/apps/{id}", Handler: appH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/apps/{id}", Handler: appH.Delete})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/apps/{id}/reset-secret", Handler: appH.ResetSecret})
	// 应用配额使用查询
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/apps/{id}/quota-usage", Handler: quotaH.GetUsage})

	// 通道管理
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/channels", Handler: chH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels", Handler: chH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}", Handler: chH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/channels/{id}", Handler: chH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/channels/{id}", Handler: chH.Delete})
	// 通道测试发送 + 健康历史
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/channels/{id}/test", Handler: ctH.Test})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}/health-history", Handler: ctH.HealthHistory})
	// 通道-模板绑定
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}/available-templates", Handler: cbH.GetAvailableTemplates})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}/bindings", Handler: cbH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/channels/{id}/bindings", Handler: cbH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}/bindings/{binding_id}", Handler: cbH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/channels/{id}/bindings/{binding_id}", Handler: cbH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/channels/{id}/bindings/{binding_id}", Handler: cbH.Delete})
	// 通道-签名映射
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}/available-signatures", Handler: csmH.GetAvailableSignatures})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}/signature-mappings", Handler: csmH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/channels/{id}/signature-mappings", Handler: csmH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/channels/{id}/signature-mappings/{mapping_id}", Handler: csmH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/channels/{id}/signature-mappings/{mapping_id}", Handler: csmH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/channels/{id}/signature-mappings/{mapping_id}", Handler: csmH.Delete})

	// 模板管理
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/templates", Handler: tplH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/templates", Handler: tplH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/templates/{id}", Handler: tplH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/templates/{id}", Handler: tplH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/templates/{id}", Handler: tplH.Delete})

	// 服务商账号管理
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-accounts/available", Handler: paH.GetAvailableProviders})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-config-fields/{provider_code}", Handler: paH.GetProviderConfigFields})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-accounts", Handler: paH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/provider-accounts", Handler: paH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-accounts/{id}", Handler: paH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/provider-accounts/{id}", Handler: paH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/provider-accounts/{id}", Handler: paH.Delete})

	// 服务商签名管理
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-signatures", Handler: psH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/provider-signatures", Handler: psH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-signatures/{id}", Handler: psH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/provider-signatures/{id}", Handler: psH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/provider-signatures/{id}", Handler: psH.Delete})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-accounts/{id}/signatures", Handler: psH.GetAvailableByProvider})

	// 供应商模板管理
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-templates", Handler: ptH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/provider-templates", Handler: ptH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-templates/{id}", Handler: ptH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/provider-templates/{id}", Handler: ptH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/provider-templates/{id}", Handler: ptH.Delete})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/provider-accounts/{id}/templates", Handler: ptH.GetAvailableByProvider})

	// 失败规则管理
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/failure-rules/options", Handler: frH.GetOptions})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/failure-rules/refresh-cache", Handler: frH.RefreshCache})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/failure-rules", Handler: frH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/failure-rules", Handler: frH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/failure-rules/{id}", Handler: frH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/failure-rules/{id}", Handler: frH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/failure-rules/{id}", Handler: frH.Delete})

	// Webhook 配置管理
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/webhook-configs", Handler: whH.List})
	protected.AddRoute(httpx.Route{Method: http.MethodPost, Path: "/webhook-configs", Handler: whH.Create})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/webhook-configs/{id}", Handler: whH.Get})
	protected.AddRoute(httpx.Route{Method: http.MethodPut, Path: "/webhook-configs/{id}", Handler: whH.Update})
	protected.AddRoute(httpx.Route{Method: http.MethodDelete, Path: "/webhook-configs/{id}", Handler: whH.Delete})

	// Webhook 日志查询
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/webhook-logs", Handler: whH.ListLogs})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/webhook-logs/task/{task_id}", Handler: whH.ListLogsByTask})

	// 回调日志查询
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/callbacks", Handler: logH.ListCallbacks})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/callbacks/task/{task_id}", Handler: logH.ListCallbacksByTask})

	// 日志查询
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/logs", Handler: logH.ListPushLogs})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/logs/task/{task_id}", Handler: logH.ListPushLogsByTask})

	// 任务查询
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/push-tasks", Handler: taskH.ListTasks})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/push-tasks/{id}", Handler: taskH.GetTask})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/push-tasks/no/{task_no}", Handler: taskH.GetTaskByNo})

	// 批量任务查询
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/batch-tasks", Handler: taskH.ListBatchTasks})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/batch-tasks/{id}", Handler: taskH.GetBatchTask})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/batch-tasks/batch/{batch_id}/tasks", Handler: taskH.ListTasksByBatch})

	// 统计分析
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/statistics", Handler: statH.GetStatistics})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/statistics/dashboard", Handler: statH.GetDashboard})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/statistics/top-applications", Handler: statH.GetTopApplications})
	protected.AddRoute(httpx.Route{Method: http.MethodGet, Path: "/statistics/recent-activities", Handler: statH.GetRecentActivities})
}
