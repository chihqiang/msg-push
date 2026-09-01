// 单条消息完整链路测试：
// 管理端配置（通道/模板/服务商/绑定）→ 应用端提交单条消息 → Worker 消费
// → 选通道/渲染 → 测试模式模拟成功 → 终态可查询 + push_log 落库。
package e2e_test

import (
	"encoding/json"
	"testing"

	"chihqiang/msg-push/model"

	"github.com/chihqiang/infra-go/httpx"
)

// TestE2E_SendSingleMessageChain 单条消息完整链路。
func TestE2E_SendSingleMessageChain(t *testing.T) {
	env := setupE2E(t)
	env.setupProviderChain(t)

	// 1. 应用端提交消息（种子应用 test-app 为测试模式，走完整链路但模拟成功）
	_, body := env.doJSON("POST", "/api/v1/messages", map[string]any{
		"channel_code":    e2eChannelCode,
		"template_code":   e2eTplCode,
		"receiver":        "13800000001",
		"template_params": map[string]string{"name": "张三", "code": "123456"},
	}, appHeaders())
	var send struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
		IsTest bool   `json:"is_test"`
	}
	if err := json.Unmarshal(parseData(t, body), &send); err != nil {
		t.Fatalf("parse send resp: %v", err)
	}
	if send.TaskID == "" {
		t.Fatalf("send returned empty task_id")
	}
	if !send.IsTest {
		t.Fatalf("seed app should be test mode, got is_test=false")
	}

	// 2. 轮询任务状态直到 success（Worker 消费 → 选通道/渲染 → 测试模式模拟成功）
	var task struct {
		Status string `json:"status"`
		IsTest bool   `json:"is_test"`
	}
	waitFor(t, "task "+send.TaskID+" to reach success", func() (bool, string) {
		_, b := env.doJSON("GET", "/api/v1/tasks/"+send.TaskID, nil, appHeaders())
		var r apiResp
		if json.Unmarshal(b, &r) != nil || r.Code != httpx.CodeOK {
			return false, string(b)
		}
		if err := json.Unmarshal(r.Data, &task); err != nil {
			return false, string(b)
		}
		return task.Status == string(model.PushTaskStatusSuccess), task.Status
	})
	if task.Status != string(model.PushTaskStatusSuccess) {
		t.Fatalf("task status = %s, want success", task.Status)
	}
	if !task.IsTest {
		t.Fatalf("task is_test = false, want true")
	}

	// 3. DB 断言：push_log 已写入且 provider_msg_id=TEST（测试模式标记）
	var logs []model.PushLog
	if err := env.db.Where("task_no = ?", send.TaskID).Find(&logs).Error; err != nil {
		t.Fatalf("query push_log: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("push_log count = %d, want 1", len(logs))
	}
	if logs[0].ProviderMsgID != "TEST" {
		t.Fatalf("provider_msg_id = %q, want TEST", logs[0].ProviderMsgID)
	}
	if !logs[0].IsTest {
		t.Fatalf("push_log is_test = false, want true")
	}
	t.Logf("PASS: 单条消息链路 OK (task=%s status=%s, push_log provider_msg_id=%s)",
		send.TaskID, task.Status, logs[0].ProviderMsgID)
}
