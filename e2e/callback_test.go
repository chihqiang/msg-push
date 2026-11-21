// 短信回执回调完整链路测试：
// 构造已发送（sending）的短信任务 + 发送日志 → 模拟阿里云回执回调
// → 定制解析器解析 → 落库 CallbackLog → 按 provider_msg_id+receiver
// 关联回填任务/日志终态 → 返回服务商期望响应（避免重复推送）。
package e2e_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"chihqiang/msg-push/model"
)

// TestE2E_SMSCallbackChain 短信回执回调链路。
func TestE2E_SMSCallbackChain(t *testing.T) {
	env := setupE2E(t)

	// 1. 创建短信服务商账号（阿里云），回调按 provider_account_id 路由
	token := env.adminToken(t)
	_, body := env.doJSON("POST", "/api/v1/account/provider-accounts", map[string]any{
		"account_name":  "e2e 阿里云短信",
		"provider_code": "aliyun_sms",
		"config": map[string]any{
			"access_key_id":     "e2e-ak",
			"access_key_secret": "e2e-sk",
		},
	}, adminHeaders(token))
	var pa idResp
	if err := json.Unmarshal(parseData(t, body), &pa); err != nil {
		t.Fatalf("parse provider account resp: %v", err)
	}
	if pa.ID == 0 {
		t.Fatalf("create provider account returned id=0")
	}

	// 2. 构造一条已发送（sending）的短信任务 + 发送日志（模拟 worker 已投递、等待回执）
	taskID := "e2e_cb_001"
	receiver := "13800000001"
	providerMsgID := "biz_e2e_cb"
	now := time.Now()
	task := model.PushTask{
		TaskID:            taskID,
		RequestID:         "e2e-request-id",
		AppID:             1, // 种子 test-app
		ChannelID:         1,
		TemplateID:        1,
		MessageType:       string(model.ChannelTypeSMS),
		Receiver:          receiver,
		Status:            model.PushTaskStatusSending,
		ProviderAccountID: pa.ID,
		SentAt:            &now,
	}
	if err := env.db.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := env.db.Create(&model.PushLog{
		RequestID:         "e2e-request-id",
		TaskID:            task.ID,
		TaskNo:            taskID,
		AppID:             task.AppID,
		ProviderAccountID: pa.ID,
		ProviderMsgID:     providerMsgID,
		Receiver:          receiver,
		Status:            string(model.PushTaskStatusSending),
		ProviderResp:      "{}",
	}).Error; err != nil {
		t.Fatalf("create push_log: %v", err)
	}

	// 3. 模拟阿里云回执回调（成功送达）
	_, cbBody := env.doJSON("POST", fmt.Sprintf("/api/callback/%d", pa.ID),
		[]map[string]any{{
			"phone_number": receiver,
			"success":      true,
			"err_code":     "DELIVRD",
			"err_msg":      "",
			"biz_id":       providerMsgID,
			"report_time":  "2026-08-16 12:00:00",
		}}, map[string]string{"Content-Type": "application/json"})

	// 4. 回调日志已落库
	var cbLogs []model.CallbackLog
	if err := env.db.Where("provider_account_id = ?", pa.ID).Find(&cbLogs).Error; err != nil {
		t.Fatalf("query callback log: %v", err)
	}
	if len(cbLogs) != 1 {
		t.Fatalf("callback_log count = %d, want 1", len(cbLogs))
	}

	// 5. 任务与发送日志被回填终态 success
	var updated model.PushTask
	if err := env.db.Where("task_id = ?", taskID).First(&updated).Error; err != nil {
		t.Fatalf("query updated task: %v", err)
	}
	if updated.Status != model.PushTaskStatusSuccess {
		t.Fatalf("task status = %s, want success (callback=%s)", updated.Status, updated.CallbackStatus)
	}
	if updated.CallbackStatus != "success" {
		t.Fatalf("task callback_status = %s, want success", updated.CallbackStatus)
	}
	if updated.CallbackTime == nil {
		t.Fatalf("task callback_time not set")
	}
	var updatedLog model.PushLog
	if err := env.db.Where("task_no = ?", taskID).First(&updatedLog).Error; err != nil {
		t.Fatalf("query updated push_log: %v", err)
	}
	if updatedLog.Status != string(model.PushTaskStatusSuccess) {
		t.Fatalf("push_log status = %s, want success", updatedLog.Status)
	}

	// 6. 回调返回服务商期望的成功响应（避免重复推送）
	var cbResp apiResp
	if err := json.Unmarshal(cbBody, &cbResp); err != nil {
		t.Fatalf("callback resp invalid: %s", cbBody)
	}
	if cbResp.Code != 0 {
		t.Fatalf("callback resp code = %d, want 0 (body=%s)", cbResp.Code, cbBody)
	}
	t.Logf("PASS: 短信回执回调链路 OK (task=%s -> %s, callback_status=%s)", taskID, updated.Status, updated.CallbackStatus)
}
