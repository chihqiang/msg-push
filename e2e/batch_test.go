// 批量消息完整链路测试：
// 应用端提交批量消息 → 整批入队 msg:batch → Worker 聚合处理
// （测试批次逐条模拟成功）→ 批次 completed 且计数正确。
package e2e_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"chihqiang/msg-push/model"
)

// TestE2E_SendBatchMessageChain 批量消息完整链路。
func TestE2E_SendBatchMessageChain(t *testing.T) {
	env := setupE2E(t)
	env.setupProviderChain(t)

	receivers := []string{"13800000001", "13800000002", "13800000003"}

	// 1. 应用端提交批量消息
	_, body := env.doJSON("POST", "/api/v1/messages/batch", map[string]any{
		"channel_code":    e2eChannelCode,
		"template_code":   e2eTplCode,
		"receivers":       receivers,
		"template_params": map[string]string{"name": "李四", "code": "888888"},
	}, appHeaders())
	var batch struct {
		BatchID string `json:"batch_id"`
	}
	if err := json.Unmarshal(parseData(t, body), &batch); err != nil {
		t.Fatalf("parse batch resp: %v", err)
	}
	if batch.BatchID == "" {
		t.Fatalf("batch send returned empty batch_id")
	}

	// 2. 轮询批次所有子任务终态（测试批次逐条模拟成功）
	waitFor(t, "batch "+batch.BatchID+" tasks to reach success", func() (bool, string) {
		var cnt int64
		env.db.Model(&model.PushTask{}).
			Where("batch_id = ? AND status = ?", batch.BatchID, string(model.PushTaskStatusSuccess)).
			Count(&cnt)
		if cnt == int64(len(receivers)) {
			return true, ""
		}
		var total int64
		env.db.Model(&model.PushTask{}).Where("batch_id = ?", batch.BatchID).Count(&total)
		return false, fmt.Sprintf("success=%d/%d", cnt, total)
	})

	// 3. DB 断言：批次状态 completed 且计数正确
	var bt model.PushBatchTask
	if err := env.db.Where("batch_id = ?", batch.BatchID).First(&bt).Error; err != nil {
		t.Fatalf("query batch task: %v", err)
	}
	if bt.Status != model.PushBatchStatusCompleted {
		t.Fatalf("batch status = %s, want completed", bt.Status)
	}
	if bt.SuccessCount != len(receivers) {
		t.Fatalf("batch success_count = %d, want %d", bt.SuccessCount, len(receivers))
	}
	if bt.FailedCount != 0 {
		t.Fatalf("batch failed_count = %d, want 0", bt.FailedCount)
	}

	// 4. 子任务全部 success 且 is_test=true
	var total, success int64
	env.db.Model(&model.PushTask{}).Where("batch_id = ?", batch.BatchID).Count(&total)
	env.db.Model(&model.PushTask{}).
		Where("batch_id = ? AND status = ? AND is_test = ?", batch.BatchID,
			string(model.PushTaskStatusSuccess), true).Count(&success)
	if total != int64(len(receivers)) || success != int64(len(receivers)) {
		t.Fatalf("tasks total=%d success(test)=%d, want %d/%d", total, success, len(receivers), len(receivers))
	}
	t.Logf("PASS: 批量消息链路 OK (batch=%s receivers=%d success=%d)", batch.BatchID, len(receivers), bt.SuccessCount)
}
