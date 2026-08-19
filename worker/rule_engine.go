package worker

import (
	"context"
	"strings"
	"sync"

	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
)

// EvaluateRequest 失败规则评估请求。
type EvaluateRequest struct {
	Scene        string          // 场景 send_failure/callback_failure
	ProviderCode string          // 服务商代码
	MessageType  string          // 消息类型
	ErrorCode    string          // 错误码
	ErrorMessage string          // 错误消息
	Task         *model.PushTask // 任务
}

// EvaluateResult 失败规则评估结果。
type EvaluateResult struct {
	Action      string             // retry/switch_provider/fail/alert
	MatchedRule *model.FailureRule // 命中的规则
	HasMatch    bool
}

// Engine 失败规则引擎接口。
type Engine interface {
	Evaluate(ctx context.Context, req *EvaluateRequest) *EvaluateResult
	RefreshCache()
}

// ruleEngine 失败规则引擎实现（内存缓存规则）。
type ruleEngine struct {
	svc   *svc.ServiceContext
	mu    sync.RWMutex
	rules map[string][]*model.FailureRule // key: scene
}

// NewEngine 创建失败规则引擎。
func NewEngine(s *svc.ServiceContext) Engine {
	e := &ruleEngine{svc: s, rules: map[string][]*model.FailureRule{}}
	e.RefreshCache()
	return e
}

// RefreshCache 重新加载规则缓存。
func (e *ruleEngine) RefreshCache() {
	cache := map[string][]*model.FailureRule{}
	for _, scene := range []string{model.RuleSceneSendFailure, model.RuleSceneCallbackFailure} {
		var rules []*model.FailureRule
		err := e.svc.DB.WithContext(context.Background()).
			Where("scene = ? AND status = 1", scene).
			Order("priority DESC, id DESC").
			Find(&rules).Error
		if err != nil {
			// 表未迁移时降级为 Warn
			logger.Warnf("refresh rule cache failed (scene=%s): %v", scene, err)
			continue
		}
		cache[scene] = rules
	}
	e.mu.Lock()
	e.rules = cache
	e.mu.Unlock()
}

// Evaluate 评估规则，命中返回动作，未命中返回场景默认动作。
func (e *ruleEngine) Evaluate(ctx context.Context, req *EvaluateRequest) *EvaluateResult {
	if req == nil {
		return &EvaluateResult{Action: model.RuleActionFail}
	}
	e.mu.RLock()
	rules := e.rules[req.Scene]
	e.mu.RUnlock()

	if len(rules) == 0 {
		e.RefreshCache()
		e.mu.RLock()
		rules = e.rules[req.Scene]
		e.mu.RUnlock()
	}

	for _, rule := range rules {
		if e.matchRule(rule, req) {
			return &EvaluateResult{Action: rule.Action, MatchedRule: rule, HasMatch: true}
		}
	}
	return e.defaultResult(req.Scene)
}

// defaultResult 场景默认动作。
func (e *ruleEngine) defaultResult(scene string) *EvaluateResult {
	switch scene {
	case model.RuleSceneSendFailure:
		return &EvaluateResult{Action: model.RuleActionRetry}
	default:
		return &EvaluateResult{Action: model.RuleActionFail}
	}
}

// matchRule 规则匹配（各条件空值即通配，多条件 AND）。
func (e *ruleEngine) matchRule(rule *model.FailureRule, req *EvaluateRequest) bool {
	if rule.ProviderCode != "" && rule.ProviderCode != req.ProviderCode {
		return false
	}
	if rule.MessageType != "" && rule.MessageType != req.MessageType {
		return false
	}
	if rule.ErrorCode != "" {
		matched := false
		for _, code := range strings.Split(rule.ErrorCode, ",") {
			if strings.TrimSpace(code) == req.ErrorCode {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if rule.ErrorKeyword != "" {
		matched := false
		msgLower := strings.ToLower(req.ErrorMessage)
		for _, kw := range strings.Split(rule.ErrorKeyword, ",") {
			if strings.Contains(msgLower, strings.ToLower(strings.TrimSpace(kw))) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
