package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// 失败规则场景常量。
const (
	RuleSceneSendFailure     = "send_failure"     // 发送失败场景
	RuleSceneCallbackFailure = "callback_failure" // 回调失败场景
)

// 失败规则动作常量。
const (
	RuleActionRetry          = "retry"           // 重试
	RuleActionSwitchProvider = "switch_provider" // 切换供应商重试
	RuleActionFail           = "fail"            // 直接失败
	RuleActionAlert          = "alert"           // 告警通知
)

// FailureRule 失败处理规则。
type FailureRule struct {
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement;comment:主键ID"`
	Name         string         `json:"name" gorm:"size:100;not null;comment:规则名称"`
	Scene        string         `json:"scene" gorm:"size:20;not null;index;comment:场景 send_failure/callback_failure"`
	ProviderCode string         `json:"provider_code" gorm:"size:50;index;comment:服务商代码(空=匹配所有)"`
	MessageType  string         `json:"message_type" gorm:"size:20;index;comment:消息类型(空=匹配所有)"`
	ErrorCode    string         `json:"error_code" gorm:"size:200;comment:错误码(逗号分隔多个)"`
	ErrorKeyword string         `json:"error_keyword" gorm:"size:200;comment:错误消息关键字(模糊匹配)"`
	Action       string         `json:"action" gorm:"size:20;not null;comment:动作 retry/switch_provider/fail/alert"`
	ActionConfig string         `json:"-" gorm:"type:text;comment:动作配置(JSON)"`
	Priority     int            `json:"priority" gorm:"not null;default:0;index;comment:优先级(越大越优先)"`
	Status       int8           `json:"status" gorm:"not null;default:1;index;comment:状态 1启用 0禁用"`
	Remark       string         `json:"remark" gorm:"size:500;comment:备注"`
	CreatedAt    time.Time      `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"comment:更新时间"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index;comment:软删除时间"`
}

// TableName 指定表名。
func (FailureRule) TableName() string {
	return "msg_failure_rules"
}

// RetryActionConfig 重试动作配置。
type RetryActionConfig struct {
	MaxRetry     int `json:"max_retry"`     // 最大重试次数
	DelaySeconds int `json:"delay_seconds"` // 重试延迟秒数
	BackoffRate  int `json:"backoff_rate"`  // 退避倍率
	MaxDelay     int `json:"max_delay"`     // 最大延迟秒数
}

// SwitchProviderActionConfig 切换供应商动作配置。
type SwitchProviderActionConfig struct {
	ExcludeCurrent bool `json:"exclude_current"` // 是否排除当前供应商
	MaxRetry       int  `json:"max_retry"`       // 切换后最大重试次数
}

// AlertActionConfig 告警动作配置。
type AlertActionConfig struct {
	WebhookURL string `json:"webhook_url"` // 告警webhook地址
	AlertLevel string `json:"alert_level"` // 告警级别 info/warning/critical
}

// GetRetryConfig 获取重试配置（带默认值）。
func (r *FailureRule) GetRetryConfig() (*RetryActionConfig, error) {
	if r.ActionConfig == "" {
		return &RetryActionConfig{MaxRetry: 3, DelaySeconds: 2, BackoffRate: 2, MaxDelay: 60}, nil
	}
	var cfg RetryActionConfig
	if err := json.Unmarshal([]byte(r.ActionConfig), &cfg); err != nil {
		return nil, err
	}
	if cfg.MaxRetry == 0 {
		cfg.MaxRetry = 3
	}
	if cfg.DelaySeconds == 0 {
		cfg.DelaySeconds = 2
	}
	if cfg.BackoffRate == 0 {
		cfg.BackoffRate = 2
	}
	if cfg.MaxDelay == 0 {
		cfg.MaxDelay = 60
	}
	return &cfg, nil
}

// GetSwitchProviderConfig 获取切换供应商配置。
func (r *FailureRule) GetSwitchProviderConfig() (*SwitchProviderActionConfig, error) {
	if r.ActionConfig == "" {
		return &SwitchProviderActionConfig{ExcludeCurrent: true, MaxRetry: 1}, nil
	}
	var cfg SwitchProviderActionConfig
	if err := json.Unmarshal([]byte(r.ActionConfig), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetAlertConfig 获取告警配置。
func (r *FailureRule) GetAlertConfig() (*AlertActionConfig, error) {
	if r.ActionConfig == "" {
		return &AlertActionConfig{AlertLevel: "warning"}, nil
	}
	var cfg AlertActionConfig
	if err := json.Unmarshal([]byte(r.ActionConfig), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
