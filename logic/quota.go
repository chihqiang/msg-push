package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chihqiang/msg-push/svc"

	"github.com/chihqiang/infra-go/logger"
	"github.com/redis/go-redis/v9"
)

// defaultDailyQuota 默认每日配额上限（应用未配置时）。
const defaultDailyQuota int64 = 10000

// quotaTTL 配额键有效期（秒），考虑跨天场景设为 48 小时。
const quotaTTL = 48 * 3600

// QuotaLogic 配额管理逻辑：基于 Redis 原子计数实现按应用每日配额。
type QuotaLogic struct {
	svc *svc.ServiceContext
}

// NewQuotaLogic 创建配额逻辑。
func NewQuotaLogic(s *svc.ServiceContext) *QuotaLogic {
	return &QuotaLogic{svc: s}
}

// quotaKey 生成配额计数键：quota:{appID}:{yyyyMMdd}（按业务时区日期）。
func quotaKey(appID uint, now time.Time) string {
	return fmt.Sprintf("quota:%d:%s", appID, now.Format("20060102"))
}

// Check 原子校验并计数：未超过 dailyLimit 时计数 +1 并返回 true，超限返回 false。
// dailyLimit <= 0 表示不限制，直接放行。
func (l *QuotaLogic) Check(ctx context.Context, appID uint, dailyLimit int) (bool, error) {
	return l.CheckN(ctx, appID, dailyLimit, 1)
}

// CheckN 原子校验并一次性扣减 count 条配额：未超过 dailyLimit 时计数 +count 并返回 true，
// 否则返回 false。用于批量发送按接收者数量精确扣减，避免一条批量请求绕过每日配额。
// dailyLimit <= 0 表示不限制；count <= 0 直接放行。
func (l *QuotaLogic) CheckN(ctx context.Context, appID uint, dailyLimit, count int) (bool, error) {
	if dailyLimit <= 0 || count <= 0 {
		return true, nil
	}
	key := quotaKey(appID, time.Now())

	// Lua 原子脚本：不存在则置 count 并设 TTL；未超限则 INCRBY；否则返回 0
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local n = tonumber(ARGV[2])
		local ttl = tonumber(ARGV[3])

		local current = redis.call('GET', key)
		if current == false then
			if n > limit then return 0 end
			redis.call('SET', key, n, 'EX', ttl)
			return 1
		end

		current = tonumber(current)
		if current + n > limit then
			return 0
		end
		redis.call('INCRBY', key, n)
		return 1
	`
	result, err := l.svc.Redis.Client().Eval(ctx, script, []string{key}, dailyLimit, count, quotaTTL).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

// Increment 手动增加配额计数（发送后补偿场景）。
func (l *QuotaLogic) Increment(ctx context.Context, appID uint, count int) error {
	if count <= 0 {
		return nil
	}
	key := quotaKey(appID, time.Now())
	if _, err := l.svc.Redis.Client().IncrBy(ctx, key, int64(count)).Result(); err != nil {
		return err
	}
	_, err := l.svc.Redis.Client().Expire(ctx, key, time.Duration(quotaTTL)*time.Second).Result()
	return err
}

// GetUsage 返回今日已用量与配额上限。
func (l *QuotaLogic) GetUsage(ctx context.Context, appID uint, dailyLimit int) (used int64, limit int64, err error) {
	if dailyLimit > 0 {
		limit = int64(dailyLimit)
	} else {
		limit = defaultDailyQuota
	}

	key := quotaKey(appID, time.Now())
	usedStr, err := l.svc.Redis.Client().Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, limit, nil
		}
		logger.Errorf("get quota usage failed: %v", err)
		return 0, limit, nil
	}
	fmt.Sscanf(usedStr, "%d", &used)
	return used, limit, nil
}
