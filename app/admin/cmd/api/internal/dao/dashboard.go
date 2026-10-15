package dao

import (
	"context"
)

type OverviewRow struct {
	TodayOrderCount     int64   `db:"today_order_count"`
	TodayGmv            float64 `db:"today_gmv"`
	YesterdayOrderCount int64   `db:"yesterday_order_count"`
	YesterdayGmv        float64 `db:"yesterday_gmv"`
	TotalOrderCount     int64   `db:"total_order_count"`
	TotalGmv            float64 `db:"total_gmv"`
	WaitPayCount        int64   `db:"wait_pay_count"`
	PaidCount           int64   `db:"paid_count"`
	CancelCount         int64   `db:"cancel_count"`
}

// Overview 汇总了 4 个库里的关键指标。
type Overview struct {
	OverviewRow
	UserCount     int64
	HomestayCount int64
	OnlineCount   int64
	BusinessCount int64
	CommentCount  int64
}

func (d *Dao) DashboardOverview(ctx context.Context) (*Overview, error) {
	o := &Overview{}
	if err := d.Order.QueryRowPartialCtx(ctx, &o.OverviewRow, "select"+
		" coalesce(sum(case when create_time >= curdate() then 1 else 0 end), 0) as today_order_count,"+
		" coalesce(sum(case when create_time >= curdate() and trade_state in (1,2) then order_total_price else 0 end), 0) / 100 as today_gmv,"+
		" coalesce(sum(case when create_time >= curdate() - interval 1 day and create_time < curdate() then 1 else 0 end), 0) as yesterday_order_count,"+
		" coalesce(sum(case when create_time >= curdate() - interval 1 day and create_time < curdate() and trade_state in (1,2) then order_total_price else 0 end), 0) / 100 as yesterday_gmv,"+
		" count(*) as total_order_count,"+
		" coalesce(sum(case when trade_state in (1,2) then order_total_price else 0 end), 0) / 100 as total_gmv,"+
		" coalesce(sum(case when trade_state = 0 then 1 else 0 end), 0) as wait_pay_count,"+
		" coalesce(sum(case when trade_state in (1,2) then 1 else 0 end), 0) as paid_count,"+
		" coalesce(sum(case when trade_state = -1 then 1 else 0 end), 0) as cancel_count"+
		" from homestay_order where del_state = 0"); err != nil {
		return nil, err
	}

	if err := d.User.QueryRowPartialCtx(ctx, &o.UserCount, "select count(*) from `user` where del_state = 0"); err != nil {
		return nil, err
	}
	if err := d.Travel.QueryRowPartialCtx(ctx, &o.HomestayCount, "select count(*) from homestay where del_state = 0"); err != nil {
		return nil, err
	}
	if err := d.Travel.QueryRowPartialCtx(ctx, &o.OnlineCount, "select count(*) from homestay where del_state = 0 and row_state = 1"); err != nil {
		return nil, err
	}
	if err := d.Travel.QueryRowPartialCtx(ctx, &o.BusinessCount, "select count(*) from homestay_business where del_state = 0"); err != nil {
		return nil, err
	}
	if err := d.Travel.QueryRowPartialCtx(ctx, &o.CommentCount, "select count(*) from homestay_comment where del_state = 0"); err != nil {
		return nil, err
	}
	return o, nil
}

type TrendRow struct {
	Date       string  `db:"date"`
	OrderCount int64   `db:"order_count"`
	Gmv        float64 `db:"gmv"`
}

func (d *Dao) OrderTrend(ctx context.Context, days int64) ([]TrendRow, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	var list []TrendRow
	err := d.Order.QueryRowsPartialCtx(ctx, &list, "select"+
		" date_format(create_time, '%Y-%m-%d') as `date`,"+
		" count(*) as order_count,"+
		" coalesce(sum(case when trade_state in (1,2) then order_total_price else 0 end), 0) / 100 as gmv"+
		" from homestay_order"+
		" where del_state = 0 and create_time >= curdate() - interval ? day"+
		" group by `date` order by `date`", days-1)
	return list, err
}

type StateRow struct {
	TradeState int64 `db:"trade_state"`
	Count      int64 `db:"count"`
}

func (d *Dao) OrderStateDist(ctx context.Context) ([]StateRow, error) {
	var list []StateRow
	err := d.Order.QueryRowsPartialCtx(ctx, &list,
		"select trade_state, count(*) as count from homestay_order where del_state = 0 group by trade_state")
	return list, err
}

type cityHomestayRow struct {
	Id   int64  `db:"id"`
	City string `db:"city"`
}

type cityOrderRow struct {
	HomestayId int64   `db:"homestay_id"`
	OrderCount int64   `db:"order_count"`
	Gmv        float64 `db:"gmv"`
}

// CityStat 民宿数量 + 成交额（成交额按民宿归属城市聚合）。
type CityStat struct {
	City  string
	Count int64
	Gmv   float64
}

func (d *Dao) CityStats(ctx context.Context) ([]CityStat, error) {
	var homestays []cityHomestayRow
	if err := d.Travel.QueryRowsPartialCtx(ctx, &homestays,
		"select id, city from homestay where del_state = 0"); err != nil {
		return nil, err
	}
	var orders []cityOrderRow
	if err := d.Order.QueryRowsPartialCtx(ctx, &orders, "select homestay_id, count(*) as order_count,"+
		" coalesce(sum(case when trade_state in (1,2) then order_total_price else 0 end), 0) / 100 as gmv"+
		" from homestay_order where del_state = 0 group by homestay_id"); err != nil {
		return nil, err
	}

	cityOf := make(map[int64]string, len(homestays))
	stat := make(map[string]*CityStat, len(homestays))
	for _, h := range homestays {
		city := h.City
		if city == "" {
			city = "其他"
		}
		cityOf[h.Id] = city
		if _, ok := stat[city]; !ok {
			stat[city] = &CityStat{City: city}
		}
		stat[city].Count++
	}
	for _, o := range orders {
		city, ok := cityOf[o.HomestayId]
		if !ok {
			continue
		}
		if _, ok := stat[city]; !ok {
			stat[city] = &CityStat{City: city}
		}
		stat[city].Gmv += o.Gmv
	}

	list := make([]CityStat, 0, len(stat))
	for _, v := range stat {
		list = append(list, *v)
	}
	// 按民宿数量倒序，取前 8 个城市
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].Count > list[i].Count {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if len(list) > 8 {
		list = list[:8]
	}
	return list, nil
}
