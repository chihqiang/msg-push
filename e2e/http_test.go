package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestE2E_HTTPResponseSemantics 校验统一响应语义：
// 分页参数默认值、业务错误到真实 HTTP 状态码的映射。
func TestE2E_HTTPResponseSemantics(t *testing.T) {
	env := setupE2E(t)
	h := adminHeaders(env.adminToken(t))

	// 1. 列表不传分页参数时回退默认 page=1&page_size=20
	status, body := env.doJSON("GET", "/api/v1/account/apps", nil, h)
	if status != http.StatusOK {
		t.Fatalf("list apps status = %d, want 200 (body=%s)", status, body)
	}
	var page struct {
		List     []json.RawMessage `json:"list"`
		Total    int64             `json:"total"`
		Page     int               `json:"page"`
		PageSize int               `json:"page_size"`
	}
	if err := json.Unmarshal(parseData(t, body), &page); err != nil {
		t.Fatalf("parse list apps: %v", err)
	}
	if page.Page != 1 || page.PageSize != 20 {
		t.Fatalf("default pagination = page %d size %d, want 1/20", page.Page, page.PageSize)
	}

	// 2. 资源不存在 → 404，业务码与消息保留
	status, body = env.doJSON("GET", "/api/v1/account/apps/999999", nil, h)
	if status != http.StatusNotFound {
		t.Fatalf("missing app status = %d, want 404 (body=%s)", status, body)
	}
	var errResp apiResp
	if err := json.Unmarshal(body, &errResp); err != nil {
		t.Fatalf("parse error resp: %v", err)
	}
	if errResp.Code != http.StatusNotFound || errResp.Msg != "application not found" {
		t.Fatalf("error resp = %+v, want code=404 msg=application not found", errResp)
	}

	// 3. 缺少令牌 → 401
	status, _ = env.doJSON("GET", "/api/v1/account/apps", nil, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", status)
	}

	// 4. 请求体缺少必填字段 → 400
	status, body = env.doJSON("POST", "/api/v1/account/apps", map[string]any{"remark": "x"}, h)
	if status != http.StatusBadRequest {
		t.Fatalf("invalid body status = %d, want 400 (body=%s)", status, body)
	}

	// 5. 密码错误 → 401
	status, body = env.doJSON("POST", "/api/v1/account/auth/login",
		map[string]string{"username": adminUser, "password": "wrong-password"}, nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d, want 401 (body=%s)", status, body)
	}

	// 6. 健康检查（rest 内置路由）
	status, body = env.doJSON("GET", "/health", nil, nil)
	if status != http.StatusOK {
		t.Fatalf("health status = %d, want 200 (body=%s)", status, body)
	}
	var health struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(parseData(t, body), &health); err != nil {
		t.Fatalf("parse health: %v", err)
	}
	if health.Status != "ok" {
		t.Fatalf("health status field = %q, want ok", health.Status)
	}
}
