package config

import (
	"time"

	"github.com/chihqiang/infra-go/httpx"
	"github.com/chihqiang/infra-go/jwt"
	"github.com/chihqiang/infra-go/logger"
	"github.com/chihqiang/infra-go/orm"
	"github.com/chihqiang/infra-go/redisx"
	"github.com/chihqiang/infra-go/taskq"
)

// Config 服务全局配置，由 conf 从 config.yaml 加载并填充默认值。
type Config struct {
	// App 应用信息。
	App AppConfig `json:"app"`
	// Server HTTP 服务配置。
	Server httpx.ServerConfig `json:"server"`
	// Logger 日志配置。
	Logger logger.Config `json:"logger"`
	// DB 数据库配置。
	DB orm.Config `json:"db"`
	// Redis 缓存配置。
	Redis redisx.Config `json:"redis"`
	// JWT 商户认证配置。
	JWT jwt.Config `json:"jwt"`
	// Taskq 异步任务队列（生产者）配置。
	Taskq taskq.Config `json:"taskq"`
	// Scheduler 后台调度配置。
	Scheduler SchedulerConfig `json:"scheduler"`
	// AccountSeed 账号种子配置（仅首次启动建库时使用）。
	AccountSeed AccountSeedConfig `json:"account_seed"`
}

// AppConfig 应用信息配置。
type AppConfig struct {
	Name string `json:",default=msg-push"`
	Env  string `json:",default=dev,options=[dev,test,prod]"`
}

// SchedulerConfig 后台调度器配置（均为可选，留空用内置默认值）。
type SchedulerConfig struct {
	// SMSCallbackTimeout 短信发送后多久未收到回执视为回调超时（默认 60s）。
	SMSCallbackTimeout time.Duration `json:",optional"`
	// SMSHardTimeout 回调超时后再等多久仍无回执则转 failed 终态（默认 10m）。
	SMSHardTimeout time.Duration `json:",optional"`
	// BusinessTimezone 业务统计时区（IANA 名，如 Asia/Shanghai）；空则用服务器本地时区。
	BusinessTimezone string `json:",optional"`
}

// AccountSeedConfig 账号种子配置（敏感信息走环境变量）。
type AccountSeedConfig struct {
	Username string `json:",optional"`
	Password string `json:",optional"`
}
