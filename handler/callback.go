package handler

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

	"chihqiang/msg-push/logic"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
)

// CallbackHandler 服务商回调接收处理器（公开路由，无需认证）。
type CallbackHandler struct {
	svc   *svc.ServiceContext
	logic *logic.CallbackLogic
}

// NewCallbackHandler 创建回调接收处理器。
func NewCallbackHandler(s *svc.ServiceContext) *CallbackHandler {
	return &CallbackHandler{svc: s, logic: logic.NewCallbackLogic(s)}
}

// Handle 处理服务商回调。POST/GET /api/callback/{id}，id 为服务商账号 ID。
func (h *CallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":400,"message":"invalid id"}`))
		return
	}

	// 读取原始请求体（Body 只能读一次，保存后重置）
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":400,"message":"failed to read body"}`))
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	// 解析表单数据
	_ = r.ParseMultipartForm(32 << 20)
	formData := make(map[string]string)
	if r.PostForm != nil {
		for k, vs := range r.PostForm {
			if len(vs) > 0 {
				formData[k] = vs[0]
			}
		}
	}

	// 收集请求头与查询参数
	headers := make(map[string]string)
	for k, vs := range r.Header {
		if len(vs) > 0 {
			headers[k] = vs[0]
		}
	}
	query := make(map[string]string)
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			query[k] = vs[0]
		}
	}

	// 打印完整 body，便于排查服务商回调问题（含手机号等接收方信息）
	logger.Infof("callback received: account_id=%d method=%s body=%s",
		id, r.Method, string(rawBody))

	result := h.logic.Handle(r.Context(), &logic.CallbackRequest{
		ProviderAccountID: uint(id),
		RawBody:           rawBody,
		Headers:           headers,
		QueryParams:       query,
		FormData:          formData,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(result.StatusCode)
	w.Write([]byte(result.Body))
}
