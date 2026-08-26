package sender

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTencentTC3SignFormat(t *testing.T) {
	auth := tencentTC3Sign("AKID_test", "secret", "2026-08-15", 1786735818, []byte(`{}`))
	if !strings.HasPrefix(auth, "TC3-HMAC-SHA256 Credential=AKID_test/2026-08-15/sms/tc3_request, SignedHeaders=content-type;host, Signature=") {
		t.Errorf("unexpected auth header: %s", auth)
	}
	// 签名必须是 64 位 hex
	sig := auth[strings.LastIndex(auth, "Signature=")+len("Signature="):]
	if len(sig) != 64 {
		t.Errorf("signature len = %d, want 64: %s", len(sig), sig)
	}
	// 确定性：相同输入产生相同签名
	auth2 := tencentTC3Sign("AKID_test", "secret", "2026-08-15", 1786735818, []byte(`{}`))
	if auth != auth2 {
		t.Error("TC3 sign not deterministic")
	}
}

func TestHMACHelpers(t *testing.T) {
	// HMAC-SHA256 已知向量：RFC 4231 test case 1
	// key = 0x0b*20, data = "Hi There"
	key := make([]byte, 20)
	for i := range key {
		key[i] = 0x0b
	}
	got := hmacSHA256(key, []byte("Hi There"))
	want := "b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7"
	if hexStr(got) != want {
		t.Errorf("hmacSHA256 = %s, want %s", hexStr(got), want)
	}
	if sha256Hex([]byte("abc")) != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Error("sha256Hex(abc) mismatch")
	}
}

func TestTencentSendRequestHeaders(t *testing.T) {
	var gotAction, gotAuth, gotRegion string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAction = r.Header.Get("X-TC-Action")
		gotAuth = r.Header.Get("Authorization")
		gotRegion = r.Header.Get("X-TC-Region")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Response":{"SendStatusSet":[{"SerialNo":"s1","Code":"Ok","Message":"ok"}]}}`))
	}))
	defer srv.Close()
	tencentSmsEndpoint = srv.URL + "/"

	s := &TencentSMSSender{}
	pa := newTestPA(map[string]any{
		"secret_id": "AKID_test", "secret_key": "secret", "sdk_app_id": "1400000001",
	})
	resp, err := s.Send(context.Background(), newSMSRequest(pa, "13800138000"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("want success, got %+v", resp)
	}
	if gotAction != "SendSms" {
		t.Errorf("X-TC-Action = %s, want SendSms", gotAction)
	}
	if !strings.HasPrefix(gotAuth, "TC3-HMAC-SHA256 Credential=") {
		t.Errorf("bad Authorization: %s", gotAuth)
	}
	if gotRegion != "ap-guangzhou" {
		t.Errorf("X-TC-Region = %s, want default ap-guangzhou", gotRegion)
	}
	if gotBody["SmsSdkAppId"] != "1400000001" {
		t.Errorf("body sdk_app_id = %v", gotBody["SmsSdkAppId"])
	}
}

func TestTencentBuildParamsOrder(t *testing.T) {
	req := newSMSRequest(newTestPA(nil), "13800138000")
	req.MappedParams = map[string]string{"name": "张三", "code": "123456"}
	// 模板是 "你好 {name}"，所以只按 {name} 顺序取值
	got := tencentBuildParams(req)
	if len(got) != 1 || got[0] != "张三" {
		t.Errorf("tencentBuildParams = %v, want [张三]", got)
	}
}

func TestTencentCallbackParser(t *testing.T) {
	p := &TencentCallbackParser{}
	ctx := context.Background()

	_, results, err := p.Parse(ctx, &CallbackRequest{RawBody: []byte(`[
		{"mobile":"13800138000","report_status":"SUCCESS","errmsg":"DELIVRD","sid":"t_001"},
		{"mobile":"13900139000","report_status":"FAIL","errmsg":"BLACKLIST","description":"黑名单","sid":"t_002"}
	]`)})
	if err != nil {
		t.Fatalf("parse tencent failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}
	if results[0].Status != "delivered" || results[0].ProviderID != "t_001" {
		t.Errorf("result[0] = %+v", results[0])
	}
	if results[1].Status != "failed" || results[1].ErrorCode != "BLACKLIST" || results[1].ErrorMessage != "黑名单" {
		t.Errorf("result[1] = %+v", results[1])
	}
}
