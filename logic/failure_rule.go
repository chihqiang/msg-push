package logic

import (
	"context"
	"errors"

	"chihqiang/msg-push/dto"
	"chihqiang/msg-push/model"
	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
	"gorm.io/gorm"
)

// FailureRuleLogic 失败规则管理逻辑。
type FailureRuleLogic struct {
	svc *svc.ServiceContext
}

// NewFailureRuleLogic 创建失败规则管理逻辑。
func NewFailureRuleLogic(s *svc.ServiceContext) *FailureRuleLogic {
	return &FailureRuleLogic{svc: s}
}

// Create 创建失败规则。
func (l *FailureRuleLogic) Create(ctx context.Context, req *dto.CreateFailureRuleRequest) (*model.FailureRule, error) {
	rule := &model.FailureRule{
		Name:         req.Name,
		Scene:        req.Scene,
		ProviderCode: req.ProviderCode,
		MessageType:  req.MessageType,
		ErrorCode:    req.ErrorCode,
		ErrorKeyword: req.ErrorKeyword,
		Action:       req.Action,
		ActionConfig: req.ActionConfig,
		Priority:     req.Priority,
		Status:       1,
		Remark:       req.Remark,
	}
	if req.Status != nil {
		rule.Status = *req.Status
	}
	if err := l.svc.DB.WithContext(ctx).Create(rule).Error; err != nil {
		return nil, err
	}
	return rule, nil
}

// List 分页查询失败规则（可按场景/状态过滤，按名称搜索）。
func (l *FailureRuleLogic) List(ctx context.Context, req *dto.FailureRuleListRequest) ([]model.FailureRule, int64, error) {
	q := l.svc.DB.WithContext(ctx).Model(&model.FailureRule{})
	if req.Scene != "" {
		q = q.Where("scene = ?", req.Scene)
	}
	if req.Status != nil {
		q = q.Where("status = ?", *req.Status)
	}
	if req.Key != "" {
		like := "%" + req.Key + "%"
		q = q.Where("name LIKE ?", like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rules []model.FailureRule
	if err := q.Order("priority DESC, id DESC").Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&rules).Error; err != nil {
		return nil, 0, err
	}
	return rules, total, nil
}

// Get 获取失败规则详情。
func (l *FailureRuleLogic) Get(ctx context.Context, id uint) (*model.FailureRule, error) {
	var rule model.FailureRule
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("failure rule not found")
		}
		return nil, err
	}
	return &rule, nil
}

// Update 更新失败规则。
func (l *FailureRuleLogic) Update(ctx context.Context, id uint, req *dto.UpdateFailureRuleRequest) error {
	var rule model.FailureRule
	if err := l.svc.DB.WithContext(ctx).Where("id = ?", id).First(&rule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("failure rule not found")
		}
		return err
	}

	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Scene != "" {
		updates["scene"] = req.Scene
	}
	if req.ProviderCode != "" {
		updates["provider_code"] = req.ProviderCode
	}
	if req.MessageType != "" {
		updates["message_type"] = req.MessageType
	}
	if req.ErrorCode != "" {
		updates["error_code"] = req.ErrorCode
	}
	if req.ErrorKeyword != "" {
		updates["error_keyword"] = req.ErrorKeyword
	}
	if req.Action != "" {
		updates["action"] = req.Action
	}
	if req.ActionConfig != "" {
		updates["action_config"] = req.ActionConfig
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if len(updates) == 0 {
		return nil
	}
	return l.svc.DB.WithContext(ctx).Model(&model.FailureRule{}).
		Where("id = ?", id).Updates(updates).Error
}

// Delete 删除失败规则。
func (l *FailureRuleLogic) Delete(ctx context.Context, id uint) error {
	res := l.svc.DB.WithContext(ctx).Delete(&model.FailureRule{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("failure rule not found")
	}
	return nil
}

// GetOptions 获取失败规则选项（场景/动作）。
func (l *FailureRuleLogic) GetOptions(ctx context.Context) *dto.FailureRuleOptionsResponse {
	return &dto.FailureRuleOptionsResponse{
		Scenes: []dto.OptionItem{
			{Value: model.RuleSceneSendFailure, Label: "发送失败"},
			{Value: model.RuleSceneCallbackFailure, Label: "回调失败"},
		},
		Actions: []dto.OptionItem{
			{Value: model.RuleActionRetry, Label: "重试"},
			{Value: model.RuleActionSwitchProvider, Label: "切换供应商重试"},
			{Value: model.RuleActionFail, Label: "直接失败"},
			{Value: model.RuleActionAlert, Label: "告警通知"},
		},
	}
}

// RefreshCache 刷新规则缓存：递增 Redis 版本号标记规则变更。
// 消费端规则引擎在缓存为空时自行重载，此处版本号用于运维观测/扩展订阅。
func (l *FailureRuleLogic) RefreshCache(ctx context.Context) error {
	key := "failure_rule_version"
	_, err := l.svc.Redis.Incr(ctx, key)
	if err != nil {
		return err
	}
	logger.Infof("failure rule cache refreshed")
	return nil
}
