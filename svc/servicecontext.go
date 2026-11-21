// Package svc 依赖装配层：聚合数据库、Redis、JWT、任务队列等全局组件。
package svc

import (
	"chihqiang/msg-push/config"
	"chihqiang/msg-push/core/auth"

	"github.com/chihqiang/ten-go/core/logger"
	"github.com/chihqiang/ten-go/core/orm"
	"github.com/chihqiang/ten-go/core/redisx"
	"github.com/chihqiang/ten-go/core/taskq"
	"gorm.io/gorm"
)

// ServiceContext 全局依赖容器，由 main 创建并注入各层。
type ServiceContext struct {
	Config   config.Config
	DB       *gorm.DB
	Redis    *redisx.Client
	Tokens   *auth.Tokens
	Producer *taskq.Producer
}

// NewServiceContext 按依赖顺序装配组件。
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	ctx := &ServiceContext{Config: c}

	// 数据库
	db, err := orm.New(c.DB)
	if err != nil {
		return nil, err
	}
	ctx.DB = db

	// Redis
	redisClient, err := redisx.New(c.Redis)
	if err != nil {
		_ = orm.Close(db)
		return nil, err
	}
	ctx.Redis = redisClient

	// JWT（商户认证）
	tokens, err := auth.NewTokens(c.JWT)
	if err != nil {
		_ = redisClient.Close()
		_ = orm.Close(db)
		return nil, err
	}
	ctx.Tokens = tokens

	// 异步任务队列生产者
	producer := taskq.NewProducer(taskq.Config{
		RedisAddr:       c.Taskq.RedisAddr,
		RedisPassword:   c.Taskq.RedisPassword,
		RedisDB:         c.Taskq.RedisDB,
		DefaultMaxRetry: c.Taskq.DefaultMaxRetry,
		DefaultTimeout:  c.Taskq.DefaultTimeout,
		DefaultQueue:    c.Taskq.DefaultQueue,
	})
	ctx.Producer = producer

	logger.Infof("service context created, db=%s redis=%s", c.DB.Driver, c.Redis.Addr)
	return ctx, nil
}

// Close 关闭所有连接类组件。
func (s *ServiceContext) Close() {
	if s.Producer != nil {
		_ = s.Producer.Close()
	}
	if s.Redis != nil {
		_ = s.Redis.Close()
	}
	if s.DB != nil {
		_ = orm.Close(s.DB)
	}
}
