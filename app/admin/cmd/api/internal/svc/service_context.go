package svc

import (
	"looklook/app/admin/cmd/api/internal/config"
	"looklook/app/admin/cmd/api/internal/dao"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	Dao    *dao.Dao
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		Dao: dao.New(
			sqlx.NewMysql(c.TravelDB.DataSource),
			sqlx.NewMysql(c.OrderDB.DataSource),
			sqlx.NewMysql(c.UsercenterDB.DataSource),
			sqlx.NewMysql(c.PaymentDB.DataSource),
			redis.MustNewRedis(c.Redis),
		),
	}
}
