package main

import (
	"context"
	"flag"
	"github.com/quanxiang-cloud/cabin/logger"

	"github.com/quanxiang-cloud/warden/api/restful"
	"github.com/quanxiang-cloud/warden/pkg/configs"

	"os"
	"os/signal"
	"syscall"
)

var (
	configPath = flag.String("config", "configs/config.yml", "-config 配置文件地址")
)

func main() {
	flag.Parse()
	log := logger.Logger
	err := configs.NewConfig(*configPath)
	if err != nil {
		panic(err)
	}

	if err != nil {
		panic(err)
	}
	// 启动路由
	ctx := context.Background()
	router, err := restful.NewRouter(ctx, configs.GetConfig(), log)
	if err != nil {
		panic(err)
	}
	go router.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			router.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
