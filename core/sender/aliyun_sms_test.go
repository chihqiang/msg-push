package sender

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAliyunSendRequestForm(t *testing.T) {
	var gotAction, gotSignName, gotPhone string
	var gotSignature string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotAction = r.FormValue("Action")
		gotSignName = r.FormValue("SignName")
		gotPhone = r.FormValue("PhoneNumbers")
		gotSignature = r.FormValue("Signature")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Code":"OK","Message":"OK","BizId":"biz_1"}`))
	}))
	defer srv.Close()
	aliyunSmsEndpoint = srv.URL + "/"

	s := &AliyunSMSSender{}
	pa := newTestPA(map[string]any{"access_key_id": "LTAI", "access_key_secret": "sec"})
	resp, err := s.Send(context.Background(), newSMSRequest(pa, "13800138000"))
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !resp.Success || resp.ProviderID != "biz_1" {
		t.Fatalf("want success+biz_1, got %+v", resp)
	}
	if gotAction != "SendSms" {
		t.Errorf("Action = %s, want SendSms", gotAction)
	}
	if gotSignName != "SIG_ALI" {
		t.Errorf("SignName = %s, want SIG_ALI", gotSignName)
	}
	if gotPhone != "13800138000" {
		t.Errorf("PhoneNumbers = %s", gotPhone)
	}
	if gotSignature == "" {
		t.Error("Signature missing from form")
	}
}

func TestSenderMissingConfig(t *testing.T) {
	// 缺少配置时返回 CONFIG_ERROR 而不是 error（业务失败，不 panic）
	s := &AliyunSMSSender{}
	resp, err := s.Send(context.Background(), newSMSRequest(newTestPA(nil), "13800138000"))
	if err != nil {
		t.Fatalf("should not return error, got %v", err)
	}
	if resp.Success {
		t.Error("want failure")
	}
	if resp.ErrorCode != "CONFIG_ERROR" {
		t.Errorf("want CONFIG_ERROR, got %s", resp.ErrorCode)
	}
}

func TestCanonicalizeAliyunParams(t *testing.T) {
	params := map[string]string{
		"Action":        "SendSms",
		"PhoneNum":      "138 0013 8000", // 含空格，验证 + → %20
		"Signature":     "a*b",           // 验证 * → %2A
		"Timestamp":     "2026-01-01T00:00:00Z",
		"TemplateParam": `{"code":"a b*c"}`,
	}
	got := canonicalizeAliyunParams(params)
	// 规范化后不得出现裸空格 / 未转义的 *（除 %2A）
	for i := 0; i < len(got); i++ {
		if got[i] == ' ' {
			t.Fatalf("canonicalize output contains raw space: %s", got)
		}
	}
	if !contains(got, "%20") {
		t.Errorf("expected %%20 encoding for space, got: %s", got)
	}
	if !contains(got, "%2A") {
		t.Errorf("expected %%2A encoding for *, got: %s", got)
	}
	// 关键：绝不能双重编码（%20 再次编码成 %2520）
	if contains(got, "%2520") || contains(got, "%252A") {
		t.Errorf("canonicalize output double-encoded: %s", got)
	}
}

// TestAliyunStringToSign 用阿里云官方文档示例验证 stringToSign 构造（修复双重编码 bug）。
// 官方示例：POST&%2F&AccessKeyId%3Dtestid%26Action%3DDescribeRegions%26Format%3DJSON%26...
func TestAliyunStringToSign(t *testing.T) {
	params := map[string]string{
		"AccessKeyId":      "testid",
		"Action":           "DescribeRegions",
		"Format":           "JSON",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   "3ee8c1b8-83d3-44af-a94f-4e0ad82fd6cf",
		"SignatureVersion": "1.0",
		"Timestamp":        "2016-02-23T12:46:24Z",
		"Version":          "2014-05-26",
	}
	query := canonicalizeAliyunParams(params)
	// 一次编码后的规范化串（含 Timestamp 冒号编码）
	wantQuery := "AccessKeyId=testid&Action=DescribeRegions&Format=JSON&SignatureMethod=HMAC-SHA1" +
		"&SignatureNonce=3ee8c1b8-83d3-44af-a94f-4e0ad82fd6cf&SignatureVersion=1.0" +
		"&Timestamp=2016-02-23T12%3A46%3A24Z&Version=2014-05-26"
	if query != wantQuery {
		t.Fatalf("canonicalizeAliyunParams =\n  %s\nwant:\n  %s", query, wantQuery)
	}
	// stringToSign = POST&%2F& + url.QueryEscape(canonicalizedQueryString)
	stringToSign := "POST&%2F&" + url.QueryEscape(query)
	wantSign := "POST&%2F&AccessKeyId%3Dtestid%26Action%3DDescribeRegions%26Format%3DJSON" +
		"%26SignatureMethod%3DHMAC-SHA1%26SignatureNonce%3D3ee8c1b8-83d3-44af-a94f-4e0ad82fd6cf" +
		"%26SignatureVersion%3D1.0%26Timestamp%3D2016-02-23T12%253A46%253A24Z%26Version%3D2014-05-26"
	if stringToSign != wantSign {
		t.Fatalf("stringToSign =\n  %s\nwant:\n  %s", stringToSign, wantSign)
	}
}

func TestAliyunCallbackParser(t *testing.T) {
	p := &AliyunCallbackParser{}
	ctx := context.Background()

	// 成功回执
	resp, results, err := p.Parse(ctx, &CallbackRequest{RawBody: []byte(`[
		{"phone_number":"13800138000","success":true,"err_code":"DELIVRD","biz_id":"biz_001"}
	]`)})
	if err != nil {
		t.Fatalf("parse aliyun ok failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("want status 200, got %d", resp.StatusCode)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	if results[0].Status != "delivered" {
		t.Errorf("want delivered, got %s", results[0].Status)
	}
	if results[0].ProviderID != "biz_001" {
		t.Errorf("want biz_001, got %s", results[0].ProviderID)
	}

	// 失败回执
	_, results, err = p.Parse(ctx, &CallbackRequest{RawBody: []byte(`[
		{"phone_number":"13800138000","success":false,"err_code":"SMSSEND_OVER_LIMIT","err_msg":"limit","biz_id":"biz_002"}
	]`)})
	if err != nil {
		t.Fatalf("parse aliyun fail: %v", err)
	}
	if results[0].Status != "failed" {
		t.Errorf("want failed, got %s", results[0].Status)
	}
	if results[0].ErrorCode != "SMSSEND_OVER_LIMIT" {
		t.Errorf("want error code, got %s", results[0].ErrorCode)
	}

	// 非法 JSON：仍返回 OK 响应避免服务商重复推送
	resp, results, err = p.Parse(ctx, &CallbackRequest{RawBody: []byte(`not-json`)})
	if err == nil {
		t.Error("want error for invalid json")
	}
	if resp.StatusCode != 200 {
		t.Errorf("invalid json should still return 200, got %d", resp.StatusCode)
	}
	if len(results) != 0 {
		t.Errorf("want 0 results for invalid json, got %d", len(results))
	}
}
