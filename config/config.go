package config

import (
	"time"

	"chihqiang/msg-push/core/auth"

	"github.com/chihqiang/ten-go/core/orm"
	"github.com/chihqiang/ten-go/core/redisx"
	"github.com/chihqiang/ten-go/core/taskq"
	"github.com/chihqiang/ten-go/rest"
)

// Config 服务全局配置，由 conf 从 config.yaml 加载并填充默认值。
type Config struct {
	// rest.Config HTTP 服务配置（嵌入：host/port/logger/middlewares/health 位于配置顶层）。
	rest.Config
	// Name 应用名。
	Name string `json:"name,default=msg-push"`
	// Env 运行环境。
	Env string `json:"env,default=dev,options=[dev,test,prod]"`
	// DB 数据库配置。
	DB orm.Config `json:"db"`
	// Redis 缓存配置。
	Redis redisx.Config `json:"redis"`
	// JWT 商户认证配置。
	JWT auth.Config `json:"jwt"`
	// Taskq 异步任务队列（生产者）配置。
	Taskq taskq.Config `json:"taskq"`
	// Scheduler 后台调度配置。
	Scheduler SchedulerConfig `json:"scheduler"`
	// AccountSeed 账号种子配置（仅首次启动建库时使用）。
	AccountSeed AccountSeedConfig `json:"account_seed"`
}

// SchedulerConfig 后台调度器配置（均为可选，留空用内置默认值）。
type SchedulerConfig struct {
	// SMSCallbackTimeout 短信发送后多久未收到回执视为回调超时（默认 60s）。
	SMSCallbackTimeout time.Duration `json:"sms_callback_timeout,optional"`
	// SMSHardTimeout 回调超时后再等多久仍无回执则转 failed 终态（默认 10m）。
	SMSHardTimeout time.Duration `json:"sms_hard_timeout,optional"`
	// BusinessTimezone 业务统计时区（IANA 名，如 Asia/Shanghai）；空则用服务器本地时区。
	BusinessTimezone string `json:"business_timezone,optional"`
}

// AccountSeedConfig 账号种子配置（敏感信息走环境变量）。
type AccountSeedConfig struct {
	Username string `json:"username,optional"`
	Password string `json:"password,optional"`
}
