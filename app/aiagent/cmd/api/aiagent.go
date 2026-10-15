package main

import (
	"context"
	"flag"
	"fmt"

	"looklook/app/aiagent/cmd/api/internal/config"
	"looklook/app/aiagent/cmd/api/internal/handler"
	"looklook/app/aiagent/cmd/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/aiagent.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// AI 服务是独立的 Agent 层：对 C 端和管理端各暴露一套接口，
	// 身份用两套 JWT 区分，工具集合按场景装配。
	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 无人值守任务：经营日报 + 异常巡检（Cron.Enabled=false 时自动跳过）
	ctx.Insight.Start(context.Background())

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
