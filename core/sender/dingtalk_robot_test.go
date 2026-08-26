package sender

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"chihqiang/msg-push/model"
)

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
