package dao

import (
	"context"
	"strings"
)

type CommentRow struct {
	Id            int64  `db:"id"`
	HomestayId    int64  `db:"homestay_id"`
	HomestayTitle string `db:"homestay_title"`
	UserId        int64  `db:"user_id"`
	Nickname      string `db:"nickname"`
	Avatar        string `db:"avatar"`
	Content       string `db:"content"`
	Star          string `db:"star"`
	CreateTime    int64  `db:"create_time"`
}

const commentSelect = "select" +
	" c.id, c.homestay_id, coalesce(h.title, '') as homestay_title," +
	" c.user_id, c.nickname, c.avatar, c.content, c.star," +
	" unix_timestamp(c.create_time) as create_time" +
	" from homestay_comment c left join homestay h on h.id = c.homestay_id"

func (d *Dao) ListComments(ctx context.Context, pageNo, pageSize, homestayId int64, keyword string) ([]CommentRow, int64, error) {
	var cb condBuilder
	cb.add("c.del_state = 0")
	if homestayId > 0 {
		cb.add("c.homestay_id = ?", homestayId)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		cb.add("(c.content like ? or c.nickname like ?)", like(kw), like(kw))
	}
	where, args := cb.sql()

	var total int64
	if err := d.Travel.QueryRowPartialCtx(ctx, &total,
		"select count(*) from homestay_comment c"+where, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	var list []CommentRow
	if err := d.Travel.QueryRowsPartialCtx(ctx, &list,
		commentSelect+where+" order by c.id desc"+limitClause(pageNo, pageSize), args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// SoftDeleteComment 删除评价，返回它所属的民宿 id（用于回写评价数）。
func (d *Dao) SoftDeleteComment(ctx context.Context, id int64) (int64, error) {
	var homestayId int64
	if err := d.Travel.QueryRowPartialCtx(ctx, &homestayId,
		"select homestay_id from homestay_comment where id = ? and del_state = 0 limit 1", id); err != nil {
		return 0, err
	}
	if _, err := d.Travel.ExecCtx(ctx,
		"update homestay_comment set del_state = 1, version = version + 1 where id = ? and del_state = 0", id); err != nil {
		return 0, err
	}
	return homestayId, nil
}
