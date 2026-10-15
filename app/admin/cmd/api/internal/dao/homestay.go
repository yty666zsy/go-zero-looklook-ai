package dao

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// HomestayRow 是管理端列表/详情用的扁平行（价格已换算成元、时间已换算成 unix 秒）。
type HomestayRow struct {
	Id                  int64   `db:"id"`
	Title               string  `db:"title"`
	SubTitle            string  `db:"sub_title"`
	Cover               string  `db:"cover"`
	City                string  `db:"city"`
	Address             string  `db:"address"`
	Tags                string  `db:"tags"`
	PeopleNum           int64   `db:"people_num"`
	HomestayPrice       float64 `db:"homestay_price"`
	MarketHomestayPrice float64 `db:"market_homestay_price"`
	FoodInfo            string  `db:"food_info"`
	FoodPrice           float64 `db:"food_price"`
	RowState            int64   `db:"row_state"`
	Score               float64 `db:"score"`
	CommentCount        int64   `db:"comment_count"`
	BusinessId          int64   `db:"homestay_business_id"`
	BusinessTitle       string  `db:"business_title"`
	CreateTime          int64   `db:"create_time"`
}

const homestaySelect = "select" +
	" h.id, h.title, h.sub_title, h.banner as cover," +
	" coalesce(h.city, '') as city, coalesce(h.address, '') as address, coalesce(h.tags, '') as tags," +
	" h.people_num," +
	" h.homestay_price / 100 as homestay_price, h.market_homestay_price / 100 as market_homestay_price," +
	" h.food_info, h.food_price / 100 as food_price, h.row_state," +
	" coalesce(h.score, 0) as score, coalesce(h.comment_count, 0) as comment_count," +
	" h.homestay_business_id, coalesce(b.title, '') as business_title," +
	" unix_timestamp(h.create_time) as create_time" +
	" from homestay h left join homestay_business b on b.id = h.homestay_business_id"

func (d *Dao) ListHomestays(ctx context.Context, pageNo, pageSize int64, keyword, city string, rowState int64) ([]HomestayRow, int64, error) {
	var cb condBuilder
	cb.add("h.del_state = 0")
	if kw := strings.TrimSpace(keyword); kw != "" {
		cb.add("(h.title like ? or h.sub_title like ? or h.city like ? or h.address like ?)", like(kw), like(kw), like(kw), like(kw))
	}
	if city = strings.TrimSpace(city); city != "" {
		cb.add("h.city = ?", city)
	}
	if rowState >= 0 {
		cb.add("h.row_state = ?", rowState)
	}
	where, args := cb.sql()

	var total int64
	if err := d.Travel.QueryRowPartialCtx(ctx, &total,
		"select count(*) from homestay h"+where, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	var list []HomestayRow
	if err := d.Travel.QueryRowsPartialCtx(ctx, &list,
		homestaySelect+where+" order by h.id desc"+limitClause(pageNo, pageSize), args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (d *Dao) FindHomestay(ctx context.Context, id int64) (*HomestayRow, error) {
	var h HomestayRow
	err := d.Travel.QueryRowPartialCtx(ctx, &h, homestaySelect+" where h.id = ? and h.del_state = 0 limit 1", id)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (d *Dao) UpdateHomestayState(ctx context.Context, id, rowState int64) error {
	_, err := d.Travel.ExecCtx(ctx,
		"update homestay set row_state = ?, version = version + 1 where id = ? and del_state = 0",
		rowState, id)
	return err
}

// UpdateHomestayPrice 入参单位是「元」，库里存「分」。
func (d *Dao) UpdateHomestayPrice(ctx context.Context, id int64, price, marketPrice float64) error {
	_, err := d.Travel.ExecCtx(ctx,
		"update homestay set homestay_price = ?, market_homestay_price = ?, version = version + 1"+
			" where id = ? and del_state = 0",
		int64(math.Round(price*100)), int64(math.Round(marketPrice*100)), id)
	return err
}

type CityOptionRow struct {
	City  string `db:"city"`
	Count int64  `db:"count"`
}

func (d *Dao) CityOptions(ctx context.Context) ([]CityOptionRow, error) {
	var list []CityOptionRow
	err := d.Travel.QueryRowsPartialCtx(ctx, &list, "select city, count(*) as count from homestay"+
		" where del_state = 0 and city <> '' group by city order by count desc limit 30")
	return list, err
}

// ---------- 评价数同步（后台删除评价后回写 homestay.comment_count） ----------

func (d *Dao) SyncHomestayCommentCount(ctx context.Context, homestayIds []int64) error {
	if len(homestayIds) == 0 {
		return nil
	}
	ids := make([]string, 0, len(homestayIds))
	for _, id := range homestayIds {
		ids = append(ids, strconv.FormatInt(id, 10))
	}
	_, err := d.Travel.ExecCtx(ctx, "update homestay set comment_count = ("+
		" select count(*) from homestay_comment c where c.homestay_id = homestay.id and c.del_state = 0"+
		" ) where id in ("+strings.Join(ids, ",")+")")
	return err
}
