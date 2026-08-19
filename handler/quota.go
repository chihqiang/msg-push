package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/httpx"
)

// QuotaHandler 配额查询处理器。
type QuotaHandler struct {
	svc   *svc.ServiceContext
	logic *logic.QuotaLogic
}

// NewQuotaHandler 创建配额查询处理器。
func NewQuotaHandler(s *svc.ServiceContext) *QuotaHandler {
	return &QuotaHandler{svc: s, logic: logic.NewQuotaLogic(s)}
}

// GetUsage 查询应用今日配额使用情况。
func (h *QuotaHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeBadRequest(r.Context(), w, "invalid id")
		return
	}

	// 查询应用
	var app model.Application
	if err := h.svc.DB.WithContext(r.Context()).
		Where("id = ?", id).
		First(&app).Error; err != nil {
		writeError(r.Context(), w, err)
		return
	}

	used, limit, err := h.logic.GetUsage(r.Context(), app.ID, app.DailyQuota)
	if err != nil {
		writeError(r.Context(), w, err)
		return
	}

	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	percentage := 0.0
	if limit > 0 {
		percentage = float64(used) / float64(limit) * 100
	}

	httpx.OkJSONCtx(r.Context(), w, dto.QuotaUsageResponse{
		DailyQuota:      int(limit),
		TodayUsed:       int(used),
		Remaining:       int(remaining),
		UsagePercentage: percentage,
	})
}
