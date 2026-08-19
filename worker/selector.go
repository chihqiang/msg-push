// Package channel 提供消费端通道选择器：平滑加权轮询 + 同接收者轮换 + 熔断。
package worker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ChannelNode 通道节点（带权重）。
type ChannelNode struct {
	ChannelTemplateBinding *model.ChannelTemplateBinding
	ProviderAccount        *model.ProviderAccount
	CurrentWeight          int
	EffectiveWeight        int
}

// Selector 通道选择器接口。
type Selector interface {
	// SelectWithExcludes 选择通道节点，支持排除指定服务商（规则引擎切换供应商时使用）。
	SelectWithExcludes(ctx context.Context, channelID uint, messageType string, appID, receiver string, excludeProviderIDs []uint) (*ChannelNode, error)
	// ReportSuccess 报告发送成功。
	ReportSuccess(providerAccountID uint)
	// ReportFailure 报告发送失败（熔断计数）。
	ReportFailure(providerAccountID uint)
}

// 常量。
const (
	// lastProviderTTL 同接收者上次供应商记录有效期。
	lastProviderTTL = 5 * time.Minute
	// weightStateTTL 权重状态有效期。
	weightStateTTL = 24 * time.Hour
	// failCountTTL 连续失败计数有效期。
	failCountTTL = 24 * time.Hour
	// nodeCacheTTL 进程内节点缓存有效期。
	// 管理端变更通道绑定/服务商配置后，消费端缓存自动过期重载，最多延迟该时长生效。
	// 由于 selector 实例不跨 handler/worker 共享，用 TTL 自动失效比显式 InvalidateCache 更稳健。
	nodeCacheTTL = 60 * time.Second
)

// 熔断相关配置。
var (
	// CircuitBreakThreshold 连续失败达到该次数自动禁用绑定。
	CircuitBreakThreshold = 5
)

// nodeCacheEntry 节点缓存条目（含加载时间用于 TTL 过期）。
type nodeCacheEntry struct {
	nodes    []*ChannelNode
	loadedAt time.Time
}

// ChannelSelector 通道选择器实现。
type ChannelSelector struct {
	svc     *svc.ServiceContext
	cache   map[string]*nodeCacheEntry // 进程内节点缓存（TTL 自动过期）
	cacheMu sync.Mutex
}

// NewSelector 创建通道选择器。
func NewSelector(s *svc.ServiceContext) *ChannelSelector {
	return &ChannelSelector{svc: s, cache: map[string]*nodeCacheEntry{}}
}

// lastProviderKey 同接收者上次供应商 key。
func lastProviderKey(appID, channelID, receiver string) string {
	return fmt.Sprintf("msgpush:last_provider:%s:%s:%s", appID, channelID, receiver)
}

// weightStateKey 权重状态 key。
func weightStateKey(channelID uint, bindingID uint) string {
	return fmt.Sprintf("msgpush:weight:%d:%d", channelID, bindingID)
}

// failCountKey 连续失败计数 key（按服务商账号）。
func failCountKey(providerAccountID uint) string {
	return fmt.Sprintf("msgpush:fail_count:%d", providerAccountID)
}

// SelectWithExcludes 选择通道节点。
func (s *ChannelSelector) SelectWithExcludes(ctx context.Context, channelID uint, messageType string, appID, receiver string, excludeProviderIDs []uint) (*ChannelNode, error) {
	nodes := s.getNodes(ctx, channelID, messageType)

	// 排除已排除的供应商
	nodes = filterExcluded(nodes, excludeProviderIDs)
	if len(nodes) == 0 {
		return nil, errors.New("no available provider after excluding")
	}

	// 5分钟内同接收者切换策略：剔除上次供应商
	if appID != "" && receiver != "" {
		last, err := s.svc.Redis.Client().Get(ctx, lastProviderKey(appID, strconv.FormatUint(uint64(channelID), 10), receiver)).Result()
		if err == nil && last != "" {
			lastID, _ := strconv.ParseUint(last, 10, 64)
			filtered := make([]*ChannelNode, 0, len(nodes))
			for _, n := range nodes {
				if n.ProviderAccount != nil && uint64(n.ProviderAccount.ID) != lastID {
					filtered = append(filtered, n)
				}
			}
			if len(filtered) > 0 {
				nodes = filtered
			}
		}
	}

	// 平滑加权轮询
	selected, err := s.smoothWeightedRoundRobin(ctx, nodes)
	if err != nil {
		return nil, err
	}

	// 记录本次供应商
	if selected != nil && selected.ProviderAccount != nil && appID != "" && receiver != "" {
		_ = s.svc.Redis.Client().Set(ctx, lastProviderKey(appID, strconv.FormatUint(uint64(channelID), 10), receiver),
			strconv.FormatUint(uint64(selected.ProviderAccount.ID), 10), lastProviderTTL).Err()
	}

	return selected, nil
}

// getNodes 获取通道可用节点（含缓存，TTL 过期后自动重载）。
func (s *ChannelSelector) getNodes(ctx context.Context, channelID uint, messageType string) []*ChannelNode {
	cacheKey := fmt.Sprintf("%d:%s", channelID, messageType)
	now := time.Now()
	s.cacheMu.Lock()
	if entry, ok := s.cache[cacheKey]; ok && now.Sub(entry.loadedAt) < nodeCacheTTL {
		s.cacheMu.Unlock()
		return entry.nodes
	}
	s.cacheMu.Unlock()

	nodes := s.loadNodesFromDB(ctx, channelID, messageType)
	s.cacheMu.Lock()
	s.cache[cacheKey] = &nodeCacheEntry{nodes: nodes, loadedAt: now}
	s.cacheMu.Unlock()
	return nodes
}

// loadNodesFromDB 从 DB 加载通道节点。
func (s *ChannelSelector) loadNodesFromDB(ctx context.Context, channelID uint, messageType string) []*ChannelNode {
	var bindings []model.ChannelTemplateBinding
	err := s.svc.DB.WithContext(ctx).
		Preload("ProviderTemplate.ProviderAccount").
		Where("channel_id = ? AND status = 1 AND is_active = 1", channelID).
		Order("priority ASC, weight DESC").
		Find(&bindings).Error
	if err != nil {
		logger.Errorf("load channel bindings failed: %v", err)
		return nil
	}

	nodes := make([]*ChannelNode, 0, len(bindings))
	for i := range bindings {
		b := &bindings[i]
		if b.ProviderTemplate == nil || b.ProviderTemplate.Status != 1 {
			continue
		}
		providerAccount := b.ProviderTemplate.ProviderAccount
		if providerAccount == nil || providerAccount.Status != 1 {
			continue
		}
		// 服务商类型需与请求消息类型匹配
		if providerAccount.ProviderType != "" && messageType != "" {
			if !typeMatch(providerAccount.ProviderType, messageType) {
				continue
			}
		}
		w := b.Weight
		if w <= 0 {
			w = 10
		}
		nodes = append(nodes, &ChannelNode{
			ChannelTemplateBinding: b,
			ProviderAccount:        providerAccount,
			EffectiveWeight:        w,
		})
	}
	return nodes
}

// typeMatch 通道类型与服务商类型匹配。
func typeMatch(providerType, messageType string) bool {
	// provider_type: sms/email/wechat_work/dingtalk；message_type: sms/email/wecom/dingtalk
	switch messageType {
	case "wecom":
		return providerType == "wechat_work"
	default:
		return providerType == messageType
	}
}

// filterExcluded 过滤被排除的供应商。
func filterExcluded(nodes []*ChannelNode, excludeProviderIDs []uint) []*ChannelNode {
	if len(excludeProviderIDs) == 0 {
		return nodes
	}
	excludeSet := map[uint]bool{}
	for _, id := range excludeProviderIDs {
		excludeSet[id] = true
	}
	filtered := make([]*ChannelNode, 0, len(nodes))
	for _, n := range nodes {
		if n.ProviderAccount != nil && !excludeSet[n.ProviderAccount.ID] {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

// smoothWeightedRoundRobin 平滑加权轮询（Nginx 算法），权重状态持久化到 Redis。
func (s *ChannelSelector) smoothWeightedRoundRobin(ctx context.Context, nodes []*ChannelNode) (*ChannelNode, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no available channel node")
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}

	// 从 Redis 加载当前权重
	for _, n := range nodes {
		key := weightStateKey(n.ChannelTemplateBinding.ChannelID, n.ChannelTemplateBinding.ID)
		val, err := s.svc.Redis.Client().Get(ctx, key).Int()
		if err != nil {
			n.CurrentWeight = 0
		} else {
			n.CurrentWeight = val
		}
	}

	totalWeight := 0
	var selected *ChannelNode
	for _, n := range nodes {
		n.CurrentWeight += n.EffectiveWeight
		totalWeight += n.EffectiveWeight
		if selected == nil || n.CurrentWeight > selected.CurrentWeight {
			selected = n
		}
	}
	if selected == nil {
		return nil, errors.New("failed to select channel node")
	}
	selected.CurrentWeight -= totalWeight

	// 写回 Redis
	pipe := s.svc.Redis.Client().Pipeline()
	for _, n := range nodes {
		key := weightStateKey(n.ChannelTemplateBinding.ChannelID, n.ChannelTemplateBinding.ID)
		pipe.Set(ctx, key, n.CurrentWeight, weightStateTTL)
	}
	_, _ = pipe.Exec(ctx)

	return selected, nil
}

// ReportSuccess 报告成功：清零连续失败计数。
func (s *ChannelSelector) ReportSuccess(providerAccountID uint) {
	if providerAccountID == 0 {
		return
	}
	ctx := context.Background()
	_ = s.svc.Redis.Client().Del(ctx, failCountKey(providerAccountID)).Err()
}

// ReportFailure 报告失败：累加连续失败计数，达阈值自动禁用绑定。
func (s *ChannelSelector) ReportFailure(providerAccountID uint) {
	if providerAccountID == 0 {
		return
	}
	ctx := context.Background()
	key := failCountKey(providerAccountID)
	count, err := s.svc.Redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return
	}
	_ = s.svc.Redis.Client().Expire(ctx, key, failCountTTL).Err()

	if int(count) >= CircuitBreakThreshold {
		// 查找该服务商账号的所有绑定并禁用
		var bindings []model.ChannelTemplateBinding
		if err := s.svc.DB.WithContext(ctx).
			Where("provider_id = ? AND status = 1 AND is_active = 1", providerAccountID).
			Find(&bindings).Error; err != nil {
			return
		}
		for _, b := range bindings {
			if b.AutoDisableOnFail {
				_ = s.svc.DB.WithContext(ctx).Model(&model.ChannelTemplateBinding{}).
					Where("id = ?", b.ID).Update("is_active", 0).Error
				logger.Warnf("circuit breaker disabled binding id=%d provider_id=%d after %d consecutive failures",
					b.ID, providerAccountID, count)
				// 立即失效对应通道的节点缓存，避免 TTL 窗口内仍选中已熔断的服务商
				s.InvalidateCache(b.ChannelID)
			}
		}
		_ = s.svc.Redis.Client().Del(ctx, key).Err()
	}
}

// InvalidateCache 使指定通道的节点缓存失效（配置变更后调用）。
func (s *ChannelSelector) InvalidateCache(channelID uint) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	for k := range s.cache {
		prefix := fmt.Sprintf("%d:", channelID)
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(s.cache, k)
		}
	}
}

// GetDB 返回 gorm DB（供消费端复用查询）。
func (s *ChannelSelector) GetDB() *gorm.DB {
	return s.svc.DB
}

// GetRedis 返回 Redis 客户端。
func (s *ChannelSelector) GetRedis() *redis.Client {
	return s.svc.Redis.Client()
}
