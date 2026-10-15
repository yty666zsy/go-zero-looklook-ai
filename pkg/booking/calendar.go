package booking

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ErrDateTaken 该日期段已被别的有效订单占用（数据库唯一键兜底命中）。
var ErrDateTaken = errors.New("selected dates already taken")

// ErrNoDate 入住区间展开后没有可占用的日期。
var ErrNoDate = errors.New("no occupiable date in range")

const (
	dateLayout         = "2006-01-02"
	duplicateEntryCode = 1062
)

// Calendar 房源逐日占用日历。
type Calendar struct {
	conn sqlx.SqlConn
}

func NewCalendar(conn sqlx.SqlConn) *Calendar {
	return &Calendar{conn: conn}
}

// Dates 把 [start, end) 展开成「每晚」的日期；离店当天不占用。
func Dates(start, end time.Time) []time.Time {
	s := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	e := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())
	out := make([]time.Time, 0, 4)
	for d := s; d.Before(e); d = d.AddDate(0, 0, 1) {
		out = append(out, d)
	}
	return out
}

// TakenDates 返回区间内已被占用的日期（锁内调用，用来给出「哪几晚冲突」的友好提示）。
func (c *Calendar) TakenDates(ctx context.Context, session sqlx.Session, homestayId int64, dates []time.Time) ([]time.Time, error) {
	if len(dates) == 0 {
		return nil, nil
	}
	holders := make([]string, 0, len(dates))
	args := make([]any, 0, len(dates)+1)
	args = append(args, homestayId)
	for _, d := range dates {
		holders = append(holders, "?")
		args = append(args, d.Format(dateLayout))
	}

	var rows []struct {
		OccupyDate time.Time `db:"occupy_date"`
	}
	query := "select occupy_date from homestay_calendar where homestay_id = ? and occupy_date in (" +
		strings.Join(holders, ",") + ")"
	if err := sessionOf(c, session).QueryRowsCtx(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	out := make([]time.Time, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.OccupyDate)
	}
	return out, nil
}

// Occupy 逐日占位。只要有一晚已被占用就会命中唯一键，返回 ErrDateTaken。
// 必须在事务里调用，和订单写在同一个事务里（要么都成功，要么都回滚）。
func (c *Calendar) Occupy(ctx context.Context, session sqlx.Session, homestayId int64, dates []time.Time, sn string) error {
	if len(dates) == 0 {
		return ErrNoDate
	}
	holders := make([]string, 0, len(dates))
	args := make([]any, 0, len(dates)*3)
	for _, d := range dates {
		holders = append(holders, "(?,?,?)")
		args = append(args, homestayId, d.Format(dateLayout), sn)
	}
	query := "insert into homestay_calendar (homestay_id,occupy_date,order_sn) values " + strings.Join(holders, ",")
	if _, err := sessionOf(c, session).ExecCtx(ctx, query, args...); err != nil {
		if isDuplicateEntry(err) {
			return ErrDateTaken
		}
		return err
	}
	return nil
}

// Release 释放某订单占用的全部日期（幂等）。取消 / 退款 / 过期时调用。
func (c *Calendar) Release(ctx context.Context, sn string) (int64, error) {
	if c == nil || c.conn == nil || sn == "" {
		return 0, nil
	}
	res, err := c.conn.ExecCtx(ctx, "delete from homestay_calendar where order_sn = ?", sn)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// sessionOf 允许传 nil session（走连接池），也允许传事务 session。
func sessionOf(c *Calendar, session sqlx.Session) sqlx.Session {
	if session != nil {
		return session
	}
	return c.conn
}

func isDuplicateEntry(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == duplicateEntryCode
	}
	// 驱动版本/包装方式不同时的兜底
	return strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}

func formatInt(v int64) string { return strconv.FormatInt(v, 10) }
