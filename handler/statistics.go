package handler

import (
	"net/http"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/ten-go/rest"
)

// StatisticHandler 统计分析处理器。
type StatisticHandler struct {
	svc   *svc.ServiceContext
	logic *logic.StatisticLogic
}

// NewStatisticHandler 创建统计分析处理器。
func NewStatisticHandler(s *svc.ServiceContext) *StatisticHandler {
	return &StatisticHandler{svc: s, logic: logic.NewStatisticLogic(s)}
}

// GetStatistics 获取统计数据（趋势/汇总/分布/Top）。
func (h *StatisticHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	var req dto.StatisticsRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	resp, err := h.logic.GetStatistics(r.Context(), &req)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// GetDashboard 获取仪表盘数据。
func (h *StatisticHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	resp, err := h.logic.GetDashboard(r.Context())
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// GetTopApplications 获取热门应用。
func (h *StatisticHandler) GetTopApplications(w http.ResponseWriter, r *http.Request) {
	var req dto.TopListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	resp, err := h.logic.GetTopApplications(r.Context(), req.Limit)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}

// GetRecentActivities 获取近期活动。
func (h *StatisticHandler) GetRecentActivities(w http.ResponseWriter, r *http.Request) {
	var req dto.TopListRequest
	if err := rest.BindQuery(r, &req); err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	req.Normalize()
	resp, err := h.logic.GetRecentActivities(r.Context(), req.Limit)
	if err != nil {
		_ = rest.ErrorJSONCtx(r.Context(), w, err)
		return
	}
	_ = rest.OkJSONCtx(r.Context(), w, resp)
}
