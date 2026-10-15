// Package dao 是管理端的数据访问层。
//
// 管理端和 C 端是两种完全不同的访问模式（C 端高并发、读多写少、强缓存；
// 管理端低频、需要全量筛选和跨表聚合），因此这里不走 go-zero 的 model 缓存，
// 而是直接对 4 个业务库执行 SQL；写操作后清一次 go-zero 的模型缓存即可。
package dao

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 订单状态（与 C 端保持一致）
const (
	OrderStateCancel   int64 = -1
	OrderStateWaitPay  int64 = 0
	OrderStateWaitUse  int64 = 1
	OrderStateUsed     int64 = 2
	OrderStateRefund   int64 = 3
	OrderStateExpired  int64 = 4
	OrderStateAll      int64 = -99
	HomestayRowOffline int64 = 0
	HomestayRowOnline  int64 = 1
	UserStatusDisabled int64 = 0
	UserStatusNormal   int64 = 1
)

type Dao struct {
	Travel  sqlx.SqlConn
	Order   sqlx.SqlConn
	User    sqlx.SqlConn
	Payment sqlx.SqlConn
	Cache   *redis.Redis
}

func New(travel, order, user, payment sqlx.SqlConn, rds *redis.Redis) *Dao {
	return &Dao{Travel: travel, Order: order, User: user, Payment: payment, Cache: rds}
}

// FlushCache 让 C 端立刻看到管理端的改动。
// 只删 cache:* （go-zero 模型缓存），不动 asynq:* 里的任务。
func (d *Dao) FlushCache(ctx context.Context) {
	if d.Cache == nil {
		return
	}
	keys, err := d.Cache.Keys("cache:*")
	if err != nil {
		logx.WithContext(ctx).Errorf("admin flush cache: keys err %v", err)
		return
	}
	if len(keys) == 0 {
		return
	}
	if _, err := d.Cache.Del(keys...); err != nil {
		logx.WithContext(ctx).Errorf("admin flush cache: del err %v", err)
	}
}

// ---------- 小工具 ----------

func like(kw string) string {
	return "%" + strings.TrimSpace(kw) + "%"
}

func intToStr(v int64) string {
	return strconv.FormatInt(v, 10)
}

// whereAnd 把非空条件拼成 " and xxx"，用于动态筛选。
type condBuilder struct {
	where []string
	args  []any
}

func (c *condBuilder) add(expr string, args ...any) {
	c.where = append(c.where, expr)
	c.args = append(c.args, args...)
}

func (c *condBuilder) addIf(ok bool, expr string, args ...any) {
	if ok {
		c.add(expr, args...)
	}
}

func (c *condBuilder) sql() (string, []any) {
	if len(c.where) == 0 {
		return "", c.args
	}
	return " where " + strings.Join(c.where, " and "), c.args
}

// page 归一化分页参数
func page(pageNo, pageSize int64) (int64, int64, int64) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageNo, pageSize, (pageNo - 1) * pageSize
}

func limitClause(pageNo, pageSize int64) string {
	_, _, offset := page(pageNo, pageSize)
	return fmt.Sprintf(" limit %d, %d", offset, pageSize)
}

// count 执行 count 查询
func count(ctx context.Context, conn sqlx.SqlConn, table, where string, args []any) (int64, error) {
	var n int64
	q := fmt.Sprintf("select count(*) from `%s`%s", table, where)
	if err := conn.QueryRowPartialCtx(ctx, &n, q, args...); err != nil {
		return 0, err
	}
	return n, nil
}
