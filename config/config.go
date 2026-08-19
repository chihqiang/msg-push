package config

import (
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
	// AccountSeed 账号种子配置（仅首次启动建库时使用）。
	AccountSeed AccountSeedConfig `json:"account_seed"`
}

// AppConfig 应用信息配置。
type AppConfig struct {
	Name string `json:",default=msg-push"`
	Env  string `json:",default=dev,options=[dev,test,prod]"`
}

// AccountSeedConfig 账号种子配置（敏感信息走环境变量）。
type AccountSeedConfig struct {
	Username string `json:",optional"`
	Password string `json:",optional"`
}
