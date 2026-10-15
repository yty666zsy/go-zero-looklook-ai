package dao

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// AdminUser 是后台管理员账号（admin_user 表，和 C 端 user 表完全隔离）。
type AdminUser struct {
	Id            int64     `db:"id"`
	Username      string    `db:"username"`
	Password      string    `db:"password"`
	Nickname      string    `db:"nickname"`
	Role          string    `db:"role"`
	Status        int64     `db:"status"`
	LastLoginTime time.Time `db:"last_login_time"`
	CreateTime    time.Time `db:"create_time"`
}

func (d *Dao) FindAdminByUsername(ctx context.Context, username string) (*AdminUser, error) {
	var a AdminUser
	err := d.User.QueryRowPartialCtx(ctx, &a,
		"select id, username, password, nickname, role, status, last_login_time, create_time"+
			" from admin_user where username = ? and del_state = 0 limit 1", username)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (d *Dao) FindAdminById(ctx context.Context, id int64) (*AdminUser, error) {
	var a AdminUser
	err := d.User.QueryRowPartialCtx(ctx, &a,
		"select id, username, password, nickname, role, status, last_login_time, create_time"+
			" from admin_user where id = ? and del_state = 0 limit 1", id)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (d *Dao) UpdateAdminPassword(ctx context.Context, id int64, password string) error {
	_, err := d.User.ExecCtx(ctx,
		"update admin_user set password = ?, version = version + 1 where id = ? and del_state = 0",
		password, id)
	return err
}

func (d *Dao) TouchAdminLoginTime(ctx context.Context, id int64, t time.Time) error {
	_, err := d.User.ExecCtx(ctx,
		"update admin_user set last_login_time = ? where id = ? and del_state = 0", t, id)
	return err
}
