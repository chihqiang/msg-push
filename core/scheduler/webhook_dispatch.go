package scheduler

import (
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"
)

// newWebhookDispatch 创建 Webhook 触发逻辑（供扫描器触发 webhook 通知）。
func newWebhookDispatch(s *svc.ServiceContext) *logic.WebhookLogic {
	return logic.NewWebhookLogic(s)
}
