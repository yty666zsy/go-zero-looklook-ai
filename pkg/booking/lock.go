// Package booking 提供「房源预订」的高并发保护原语。
//
// 背景：民宿是整套房源，同一房源同一晚只能卖一次。原来的下单链路是
// 「查一下 → 直接 insert」，没有任何并发控制，N 个请求同时抢同一晚
// 会全部下单成功（实测 50 并发卖出 40 单）。
//
// 这里用两层保护：
//   - Lock：Redis 分布式锁（按房源粒度），把「查占用 + 占位 + 落订单」串起来，
//     让同一房源的并发下单排队执行，避免大部分冲突打到数据库；
//   - Calendar：房源逐日占用日历，靠数据库唯一键 (homestay_id, occupy_date) 兜底，
//     即使锁失效/被绕过，也一定不会超卖（宁可下单失败，不可卖重）。
package booking

import (
	"context"
	"errors"
	"time"

	"looklook/pkg/tool"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// ErrLockBusy 抢锁超时：同一房源正被并发抢订，让调用方换个话术提示重试。
var ErrLockBusy = errors.New("booking lock busy")

// unlockScript 只删自己加的锁（比对 token），避免误删别人的锁。
// SETNX 过期时间到期后锁自动消失，所以这里不做续租。
const unlockScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
end
return 0
`

// Lock 基于 Redis 的分布式锁。
type Lock struct {
	rds *redis.Redis
}

func NewLock(rds *redis.Redis) *Lock {
	return &Lock{rds: rds}
}

// Key 房源下单锁的 key。
func LockKey(homestayId int64) string {
	return "lock:booking:homestay:" + formatInt(homestayId)
}

// Acquire 在 waitMs 毫秒内自旋抢锁，成功返回解锁 token。
// ttlSec 是锁的自动过期时间：进程崩了也不会把房源锁死。
func (l *Lock) Acquire(ctx context.Context, key string, ttlSec, waitMs int) (string, error) {
	if l == nil || l.rds == nil {
		return "", errors.New("booking lock not configured")
	}
	if ttlSec <= 0 {
		ttlSec = 5
	}
	token := tool.Krand(16, tool.KC_RAND_KIND_ALL)

	deadline := time.Now().Add(time.Duration(waitMs) * time.Millisecond)
	for {
		ok, err := l.rds.SetnxExCtx(ctx, key, token, ttlSec)
		if err != nil {
			return "", err
		}
		if ok {
			return token, nil
		}
		if time.Now().After(deadline) {
			return "", ErrLockBusy
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// Release 释放锁（幂等，token 不匹配时不做任何事）。
func (l *Lock) Release(ctx context.Context, key, token string) error {
	if l == nil || l.rds == nil || token == "" {
		return nil
	}
	_, err := l.rds.EvalCtx(ctx, unlockScript, []string{key}, token)
	return err
}
