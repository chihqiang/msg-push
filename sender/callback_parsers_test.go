package sender

import (
	"context"
	"testing"
)

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

func TestCallbackParserRegistry(t *testing.T) {
	for _, code := range []string{"aliyun_sms", "tencent_sms", "netease_sms"} {
		if p := GetCallbackParser(code); p == nil {
			t.Errorf("callback parser not registered for %s", code)
		}
	}
	if p := GetCallbackParser("smtp"); p != nil {
		t.Errorf("smtp should not have a callback parser")
	}
}
