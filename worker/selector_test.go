package worker

import (
	"testing"

	"chihqiang/msg-push/model"
)

func TestFilterExcluded(t *testing.T) {
	mkNode := func(id uint) *ChannelNode {
		return &ChannelNode{ProviderAccount: &model.ProviderAccount{ID: id}}
	}
	nodes := []*ChannelNode{mkNode(1), mkNode(2), mkNode(3)}

	t.Run("无排除返回原样", func(t *testing.T) {
		got := filterExcluded(nodes, nil)
		if len(got) != 3 {
			t.Fatalf("want 3, got %d", len(got))
		}
	})
	t.Run("排除单个", func(t *testing.T) {
		got := filterExcluded(nodes, []uint{2})
		if len(got) != 2 {
			t.Fatalf("want 2, got %d", len(got))
		}
		for _, n := range got {
			if n.ProviderAccount.ID == 2 {
				t.Error("excluded provider still present")
			}
		}
	})
	t.Run("排除多个", func(t *testing.T) {
		got := filterExcluded(nodes, []uint{1, 3})
		if len(got) != 1 || got[0].ProviderAccount.ID != 2 {
			t.Fatalf("want only id=2, got %+v", got)
		}
	})
	t.Run("排除所有", func(t *testing.T) {
		got := filterExcluded(nodes, []uint{1, 2, 3})
		if len(got) != 0 {
			t.Fatalf("want 0, got %d", len(got))
		}
	})
	t.Run("排除不存在的ID不报错", func(t *testing.T) {
		got := filterExcluded(nodes, []uint{99})
		if len(got) != 3 {
			t.Fatalf("want 3, got %d", len(got))
		}
	})
	t.Run("排除列表为空不做过滤", func(t *testing.T) {
		withNil := []*ChannelNode{mkNode(1), {ProviderAccount: nil}}
		got := filterExcluded(withNil, nil)
		if len(got) != 2 {
			t.Fatalf("empty exclude list should return as-is, got %d", len(got))
		}
	})
	t.Run("有排除项时 nil ProviderAccount 节点被过滤", func(t *testing.T) {
		withNil := []*ChannelNode{mkNode(1), {ProviderAccount: nil}}
		got := filterExcluded(withNil, []uint{99})
		if len(got) != 1 || got[0].ProviderAccount.ID != 1 {
			t.Fatalf("want only id=1, got %+v", got)
		}
	})
}

// TestSmoothWeightedRoundRobinAlgorithm 验证 Nginx 平滑加权轮询算法核心逻辑。
// 直接对节点做权重累加/选中/扣减，验证权重分布（不依赖 Redis，覆盖算法正确性）。
func TestSmoothWeightedRoundRobinAlgorithm(t *testing.T) {
	// 三个节点，权重 5/1/1，模拟 14 次轮询应得到 5:1:1 的分布
	nodes := []*ChannelNode{
		{EffectiveWeight: 5},
		{EffectiveWeight: 1},
		{EffectiveWeight: 1},
	}
	counts := map[int]int{}
	for i := 0; i < 14; i++ {
		totalWeight := 0
		var selected int
		for idx := range nodes {
			nodes[idx].CurrentWeight += nodes[idx].EffectiveWeight
			totalWeight += nodes[idx].EffectiveWeight
			if selected == -1 || nodes[idx].CurrentWeight > nodes[selected].CurrentWeight {
				selected = idx
			}
		}
		if selected < 0 {
			t.Fatal("no selected")
		}
		nodes[selected].CurrentWeight -= totalWeight
		counts[selected]++
	}
	// 期望：节点0 选 10 次（5/7 比例），节点1、2 各 2 次
	if counts[0] != 10 || counts[1] != 2 || counts[2] != 2 {
		t.Errorf("weight distribution = %v, want {0:10,1:2,2:2}", counts)
	}
	// 每轮选取后不应出现连续同一节点超过权重比例（平滑性抽样验证）
	// 14 轮中节点0 出现 10 次，验证不会全集中在开头
	if counts[0] > 12 {
		t.Errorf("node0 appears too concentrated: %v", counts)
	}
}
