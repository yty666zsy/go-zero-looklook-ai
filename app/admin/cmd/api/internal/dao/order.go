package dao

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type OrderRow struct {
	Id              int64   `db:"id"`
	Sn              string  `db:"sn"`
	UserId          int64   `db:"user_id"`
	UserMobile      string  `db:"user_mobile"`
	HomestayId      int64   `db:"homestay_id"`
	Title           string  `db:"title"`
	SubTitle        string  `db:"sub_title"`
	Cover           string  `db:"cover"`
	LiveStartDate   int64   `db:"live_start_date"`
	LiveEndDate     int64   `db:"live_end_date"`
	LivePeopleNum   int64   `db:"live_people_num"`
	NeedFood        int64   `db:"need_food"`
	HomestayPrice   float64 `db:"homestay_price"`
	FoodPrice       float64 `db:"food_price"`
	OrderTotalPrice float64 `db:"order_total_price"`
	TradeState      int64   `db:"trade_state"`
	TradeCode       string  `db:"trade_code"`
	Remark          string  `db:"remark"`
	CreateTime      int64   `db:"create_time"`
}

const orderSelect = "select" +
	" id, sn, user_id, homestay_id, title, sub_title, cover," +
	" unix_timestamp(live_start_date) as live_start_date, unix_timestamp(live_end_date) as live_end_date," +
	" live_people_num, need_food," +
	" homestay_price / 100 as homestay_price, food_price / 100 as food_price," +
	" order_total_price / 100 as order_total_price," +
	" trade_state, trade_code, remark, unix_timestamp(create_time) as create_time" +
	" from homestay_order"

func (d *Dao) ListOrders(ctx context.Context, pageNo, pageSize, tradeState int64, keyword string) ([]OrderRow, int64, float64, error) {
	var cb condBuilder
	cb.add("del_state = 0")
	if tradeState != OrderStateAll {
		cb.add("trade_state = ?", tradeState)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		cb.add("(sn like ? or title like ? or trade_code like ?)", like(kw), like(kw), like(kw))
	}
	where, args := cb.sql()

	var meta struct {
		Total int64   `db:"total"`
		Sum   float64 `db:"sum_price"`
	}
	if err := d.Order.QueryRowPartialCtx(ctx, &meta,
		"select count(*) as total, coalesce(sum(order_total_price), 0) / 100 as sum_price from homestay_order"+where,
		args...); err != nil {
		return nil, 0, 0, err
	}
	if meta.Total == 0 {
		return nil, 0, 0, nil
	}

	var list []OrderRow
	if err := d.Order.QueryRowsPartialCtx(ctx, &list,
		orderSelect+where+" order by id desc"+limitClause(pageNo, pageSize), args...); err != nil {
		return nil, 0, 0, err
	}
	return list, meta.Total, meta.Sum, nil
}

func (d *Dao) FindOrderBySn(ctx context.Context, sn string) (*OrderRow, error) {
	var o OrderRow
	err := d.Order.QueryRowPartialCtx(ctx, &o, orderSelect+" where sn = ? and del_state = 0 limit 1", sn)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// OrderCountByUserId 批量取用户的下单数。
func (d *Dao) OrderCountByUserId(ctx context.Context, userIds []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(userIds))
	if len(userIds) == 0 {
		return result, nil
	}
	ids := make([]string, 0, len(userIds))
	for _, id := range userIds {
		ids = append(ids, intToStr(id))
	}
	var rows []struct {
		UserId int64 `db:"user_id"`
		Count  int64 `db:"count"`
	}
	if err := d.Order.QueryRowsPartialCtx(ctx, &rows,
		"select user_id, count(*) as count from homestay_order where del_state = 0 and user_id in ("+
			strings.Join(ids, ",")+") group by user_id"); err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.UserId] = r.Count
	}
	return result, nil
}

// ForceCancelOrder 强制取消订单（仅 待支付/未使用 可取消）；已支付的同时标记为已退款。
func (d *Dao) ForceCancelOrder(ctx context.Context, sn, reason string) (int64, int64, error) {
	var old int64
	if err := d.Order.QueryRowPartialCtx(ctx, &old,
		"select trade_state from homestay_order where sn = ? and del_state = 0 limit 1", sn); err != nil {
		return 0, 0, err
	}
	res, err := d.Order.ExecCtx(ctx, "update homestay_order"+
		" set trade_state = ?, remark = ?, version = version + 1"+
		" where sn = ? and del_state = 0 and trade_state in (0, 1)", OrderStateCancel, reason, sn)
	if err != nil {
		return 0, old, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, old, err
	}
	if n > 0 {
		// 释放房源占用日历（与 order-rpc 的释放逻辑一致）：取消后这些日期重新可订
		if _, err := d.Order.ExecCtx(ctx, "delete from homestay_calendar where order_sn = ?", sn); err != nil {
			// 只记日志，不影响主流程
			logx.WithContext(ctx).Errorf("release homestay calendar fail, sn : %s , err : %v", sn, err)
		}
	}
	if n > 0 && old == OrderStateWaitUse {
		// 已支付 → 同步把支付流水标记为「已退款」
		if _, err := d.Payment.ExecCtx(ctx, "update third_payment"+
			" set pay_status = 2, version = version + 1"+
			" where order_sn = ? and del_state = 0 and pay_status = 1", sn); err != nil {
			return n, old, err
		}
	}
	return n, old, nil
}

// UserMobiles 批量取用户手机号（管理端展示用）。
func (d *Dao) UserMobiles(ctx context.Context, userIds []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(userIds))
	if len(userIds) == 0 {
		return result, nil
	}
	ids := make([]string, 0, len(userIds))
	for _, id := range userIds {
		ids = append(ids, intToStr(id))
	}
	var rows []struct {
		Id     int64  `db:"id"`
		Mobile string `db:"mobile"`
	}
	if err := d.User.QueryRowsPartialCtx(ctx, &rows,
		"select id, mobile from `user` where id in ("+strings.Join(ids, ",")+")"); err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.Id] = r.Mobile
	}
	return result, nil
}
