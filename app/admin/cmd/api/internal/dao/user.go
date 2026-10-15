package dao

import (
	"context"
	"strings"
)

type UserRow struct {
	Id         int64  `db:"id"`
	Mobile     string `db:"mobile"`
	Nickname   string `db:"nickname"`
	Avatar     string `db:"avatar"`
	Sex        int64  `db:"sex"`
	Info       string `db:"info"`
	Status     int64  `db:"status"`
	CreateTime int64  `db:"create_time"`
	OrderCount int64  `db:"order_count"`
}

const userSelect = "select id, mobile, nickname, avatar, sex, info, status," +
	" unix_timestamp(create_time) as create_time from `user`"

func (d *Dao) ListUsers(ctx context.Context, pageNo, pageSize int64, keyword string) ([]UserRow, int64, error) {
	var cb condBuilder
	cb.add("del_state = 0")
	if kw := strings.TrimSpace(keyword); kw != "" {
		cb.add("(mobile like ? or nickname like ?)", like(kw), like(kw))
	}
	where, args := cb.sql()

	var total int64
	if err := d.User.QueryRowPartialCtx(ctx, &total, "select count(*) from `user`"+where, args...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	var list []UserRow
	if err := d.User.QueryRowsPartialCtx(ctx, &list,
		userSelect+where+" order by id desc"+limitClause(pageNo, pageSize), args...); err != nil {
		return nil, 0, err
	}

	ids := make([]int64, 0, len(list))
	for _, u := range list {
		ids = append(ids, u.Id)
	}
	counts, err := d.OrderCountByUserId(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range list {
		list[i].OrderCount = counts[list[i].Id]
	}
	return list, total, nil
}

func (d *Dao) UpdateUserStatus(ctx context.Context, id, status int64) error {
	_, err := d.User.ExecCtx(ctx,
		"update `user` set status = ?, version = version + 1 where id = ? and del_state = 0", status, id)
	return err
}
