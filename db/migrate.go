// Package db 负责数据库迁移与种子数据。
// 按"新项目"处理：建表 + 种子数据直接填充，不做幂等兼容。
// 数据库清理（删库/清空）由运维手动完成，本代码不负责清表。
package db

import (
	"fmt"

	"chihqiang/msg-push/model"

	"github.com/chihqiang/infra-go/hash"
	"github.com/chihqiang/infra-go/logger"
	"gorm.io/gorm"
)

// Migrate 自动建表。数据库需为全新/已手动清理的状态，否则会与旧数据冲突。
func Migrate(gormDB *gorm.DB) error {
	return gormDB.AutoMigrate(
		&model.Application{},
		&model.Channel{},
		&model.MessageTemplate{},
		&model.PushTask{},
		&model.PushLog{},
		&model.Account{},
		&model.ProviderAccount{},
		&model.ProviderSignature{},
		&model.ProviderTemplate{},
		&model.ChannelTemplateBinding{},
		&model.ChannelSignatureMapping{},
		&model.ChannelHealthHistory{},
		&model.FailureRule{},
		&model.CallbackLog{},
		&model.WebhookConfig{},
		&model.WebhookLog{},
		&model.PushBatchTask{},
		&model.AppQuotaStat{},
		&model.ProviderQuotaStat{},
	)
}

// Seed 填充种子数据：默认应用、账号种子（表为全新状态，直接创建）。
func Seed(gormDB *gorm.DB, accountUsername, accountPassword string) error {
	if err := seedDemoApp(gormDB); err != nil {
		return err
	}
	if err := seedAccount(gormDB, accountUsername, accountPassword); err != nil {
		return err
	}
	return nil
}

// seedDemoApp 创建演示应用。
func seedDemoApp(gormDB *gorm.DB) error {
	secretHash, err := hash.BcryptHashDefault("test-secret")
	if err != nil {
		return fmt.Errorf("seed demo app hash: %w", err)
	}
	app := model.Application{
		AppID:          "test-app",
		AppSecret:      secretHash,
		AppSecretPlain: "test-secret",
		Name:           "测试应用",
		Status:         1,
		// 演示应用默认测试模式：走完整链路但不真实发送（模拟成功）
		IsTest:     true,
		DailyQuota: 10000,
		RateLimit:  100,
		Remark:     "内置测试应用",
	}
	if err := gormDB.Create(&app).Error; err != nil {
		return fmt.Errorf("seed demo app: %w", err)
	}
	logger.Infof("seeded test app: app_id=test-app")
	return nil
}

// seedAccount 创建账号种子。
func seedAccount(gormDB *gorm.DB, username, password string) error {
	if username == "" || password == "" {
		logger.Warn("account seed skipped: empty username/password")
		return nil
	}
	passwordHash, err := hash.BcryptHashDefault(password)
	if err != nil {
		return fmt.Errorf("seed account hash: %w", err)
	}
	account := model.Account{
		Username: username,
		Password: passwordHash,
		Name:     "账号",
		Status:   1,
	}
	if err := gormDB.Create(&account).Error; err != nil {
		return fmt.Errorf("seed account: %w", err)
	}
	logger.Infof("seeded account: %s", username)
	return nil
}
