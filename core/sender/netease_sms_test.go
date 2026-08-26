package sender

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNeteaseSendChecksum(t *testing.T) {
	var gotAppKey, gotNonce, gotCurTime, gotChecksum string
	var gotMobile string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotAppKey = r.Header.Get("AppKey")
		gotNonce = r.Header.Get("Nonce")
		gotCurTime = r.Header.Get("CurTime")
		gotChecksum = r.Header.Get("CheckSum")
		gotMobile = r.FormValue("mobiles")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":200,"msg":"ok","obj":"sendid_1"}`))
	}))
	defer srv.Close()
	neteaseSendTemplateURL = srv.URL + "/"

	s := &NeteaseSMSSender{}
	pa := newTestPA(map[string]any{"app_key": "appk", "app_secret": "appsec"})
	resp, err := s.Send(context.Background(), newSMSRequest(pa, "13800138000"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !resp.Success || resp.ProviderID != "sendid_1" {
		t.Fatalf("want success, got %+v", resp)
	}
	if gotAppKey != "appk" {
		t.Errorf("AppKey = %s", gotAppKey)
	}
	// checksum = sha1(appSecret + nonce + curTime)
	expected := hexStr(sha1Sum([]byte("appsec" + gotNonce + gotCurTime)))
	if gotChecksum != expected {
		t.Errorf("CheckSum = %s, want %s", gotChecksum, expected)
	}
	if !strings.Contains(gotMobile, "13800138000") {
		t.Errorf("mobiles = %s", gotMobile)
	}
}

func TestNeteaseFlexString(t *testing.T) {
	if got := neteaseFlexString("123"); got != "123" {
		t.Errorf("neteaseFlexString(string)=%q", got)
	}
	if got := neteaseFlexString(float64(123)); got != "123" {
		t.Errorf("neteaseFlexString(float)=%q, want 123", got)
	}
	if got := neteaseFlexString(nil); got != "" {
		t.Errorf("neteaseFlexString(nil)=%q, want empty", got)
	}
}

func TestNeteaseCallbackParser(t *testing.T) {
	p := &NeteaseCallbackParser{}
	ctx := context.Background()

	// 下行回执 eventType=11
	resp, results, err := p.Parse(ctx, &CallbackRequest{RawBody: []byte(`{
		"eventType":"11",
		"objects":[{"mobile":"13800138000","sendid":"1490","result":"DELIVRD","reason":"","reportTime":"2026-01-01 10:00:00"}]
	}`)})
	if err != nil {
		t.Fatalf("parse netease report failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].Type != "report" || results[0].Status != "delivered" || results[0].ProviderID != "1490" {
		t.Errorf("report result = %+v", results[0])
	}

	// 上行短信 eventType=12
	_, results, err = p.Parse(ctx, &CallbackRequest{RawBody: []byte(`{
		"eventType":"12",
		"objects":[{"mobile":"13800138000","content":"TD","receiveTime":"2026-01-01 10:05:00"}]
	}`)})
	if err != nil {
		t.Fatalf("parse netease upstream failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 upstream, got %d", len(results))
	}
	if results[0].Type != "upstream" || results[0].Content != "TD" {
		t.Errorf("upstream result = %+v", results[0])
	}

	// 未知事件：空结果但无错误
	_, results, err = p.Parse(ctx, &CallbackRequest{RawBody: []byte(`{"eventType":"99","objects":[]}`)})
	if err != nil {
		t.Fatalf("parse netease unknown event: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("want 0 results for unknown event, got %d", len(results))
	}
}
