package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	// 管理端独立 JWT
	JwtAuth struct {
		AccessSecret string
		AccessExpire int64
	}

	Redis redis.RedisConf

	// 直接连业务库（管理端不经过 RPC）
	TravelDB struct {
		DataSource string
	}
	OrderDB struct {
		DataSource string
	}
	PaymentDB struct {
		DataSource string
	}
	UsercenterDB struct {
		DataSource string
	}
}
