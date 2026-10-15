package svc

import (
	"github.com/hibiken/asynq"
	"looklook/app/order/cmd/rpc/internal/config"
	"looklook/app/order/model"
	"looklook/app/travel/cmd/rpc/travel"
	"looklook/pkg/booking"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	AsynqClient *asynq.Client

	TravelRpc travel.Travel

	HomestayOrderModel model.HomestayOrderModel

	// 高并发防超卖：房源粒度分布式锁 + 逐日占用日历（见 pkg/booking）
	DB       sqlx.SqlConn
	Lock     *booking.Lock
	Calendar *booking.Calendar
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DB.DataSource)
	return &ServiceContext{
		Config: c,
		AsynqClient:newAsynqClient(c),

		TravelRpc: travel.NewTravel(zrpc.MustNewClient(c.TravelRpcConf)),

		HomestayOrderModel: model.NewHomestayOrderModel(conn, c.Cache),

		DB:       conn,
		Lock:     booking.NewLock(redis.MustNewRedis(c.Redis.RedisConf)),
		Calendar: booking.NewCalendar(conn),
	}
}
