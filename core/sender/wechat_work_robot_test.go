package sender

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chihqiang/msg-push/model"
)

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
