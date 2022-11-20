package main

import (
	"context"
	"distributedDemo/logger"
	"distributedDemo/registry"
	"distributedDemo/service"
	"fmt"
	"log"
)

func main() {
	//todo 将各种配置使用配置文件的方式读取
	logger.Run("./distributed.log")

	host, port := "localhost", ":4000"
	serviceAddress := fmt.Sprintf("http://%s%s", host, port)

	r := registry.Registration{
		ServiceName:      registry.LoggerService,
		ServiceURL:       serviceAddress,
		RequiredServices: make([]registry.ServiceName, 0),
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat",
	}
	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		logger.RegisterHandlers,
	)
	if err != nil {
		log.Fatalln("starting", registry.LoggerService, ":", err)
	}
	//服务启动失败或手动终止时
	<-ctx.Done()
	fmt.Println("Shutting down logger service")
}
