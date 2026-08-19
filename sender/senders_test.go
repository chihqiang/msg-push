package sender

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"chihqiang/msg-push/model"
)

func newTestPA(cfg map[string]any) *model.ProviderAccount {
	pa := &model.ProviderAccount{}
	_ = pa.SetConfig(cfg)
	return pa
}

func newSMSRequest(pa *model.ProviderAccount, receiver string) *SendRequest {
	return &SendRequest{
		Task:            &model.PushTask{TaskID: "task_t1", Receiver: receiver},
		ProviderAccount: pa,
		ChannelTemplateBinding: &model.ChannelTemplateBinding{
			ProviderTemplate: &model.ProviderTemplate{TemplateCode: "TPL_001", TemplateContent: "你好 {name}"},
		},
		Signature:    &model.ProviderSignature{SignatureCode: "SIG_ALI"},
		MappedParams: map[string]string{"name": "张三"},
	}
}

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

func hexStr(b []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = digits[v>>4]
		out[i*2+1] = digits[v&0xf]
	}
	return string(out)
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

func TestWeChatWorkRobotSend(t *testing.T) {
	var gotURL, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		b := new(strings.Builder)
		_, _ = copyBuffer(b, r.Body)
		gotBody = b.String()
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	s := &WeChatWorkRobotSender{}
	pa := newTestPA(map[string]any{"webhook_url": srv.URL + "/cgi-bin/webhook/send", "msg_type": "text"})
	resp, err := s.Send(context.Background(), &SendRequest{
		Task:            &model.PushTask{TaskID: "t_robot", Receiver: "group"},
		ProviderAccount: pa,
		RenderedContent: "测试机器人消息",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("want success, got %+v", resp)
	}
	if !strings.Contains(gotURL, "/cgi-bin/webhook/send") {
		t.Errorf("url = %s", gotURL)
	}
	var payload map[string]any
	_ = json.Unmarshal([]byte(gotBody), &payload)
	if payload["msgtype"] != "text" {
		t.Errorf("msgtype = %v", payload["msgtype"])
	}
}

func TestDingTalkRobotSendSign(t *testing.T) {
	var gotQuery string
	var gotTimestamp, gotSign string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		gotTimestamp = r.URL.Query().Get("timestamp")
		gotSign = r.URL.Query().Get("sign")
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer srv.Close()

	s := &DingTalkRobotSender{}
	pa := newTestPA(map[string]any{
		"webhook_url": srv.URL + "/robot/send", "secret": "SEC123",
	})
	resp, err := s.Send(context.Background(), &SendRequest{
		Task:            &model.PushTask{TaskID: "t_dt", Receiver: "group"},
		ProviderAccount: pa,
		RenderedContent: "钉钉测试",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if !resp.Success {
		t.Fatalf("want success, got %+v", resp)
	}
	if gotTimestamp == "" || gotSign == "" {
		t.Fatalf("missing sign params, query=%s", gotQuery)
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

func TestNormalizeEmailSubject(t *testing.T) {
	if got := normalizeEmailSubject(""); got != "通知" {
		t.Errorf("empty -> %q, want 通知", got)
	}
	if got := normalizeEmailSubject("  验证码  "); got != "验证码" {
		t.Errorf("trim -> %q", got)
	}
	if got := normalizeEmailSubject("标题\r\n注入"); got != "通知" {
		t.Errorf("with CRLF should fallback, got %q", got)
	}
	if got := normalizeEmailSubject("正常标题"); got != "正常标题" {
		t.Errorf("normal -> %q", got)
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

func TestStatusQueryRequestFields(t *testing.T) {
	now := time.Now()
	req := &StatusQueryRequest{
		Task: &model.PushTask{TaskID: "t"}, ProviderAccount: newTestPA(nil),
		ProviderMsgID: "biz", PhoneNumber: "13800138000", SendDate: now,
	}
	if req.ProviderMsgID != "biz" || req.PhoneNumber != "13800138000" {
		t.Error("StatusQueryRequest fields not set")
	}
}
