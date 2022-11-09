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
		ServiceName: "Logger Service",
		ServiceURL:  serviceAddress,
	}
	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		logger.RegisterHandlers,
	)
	if err != nil {
		log.Fatalln(err)
	}
	//服务启动失败或手动终止时
	<-ctx.Done()
	fmt.Println("Shutting down logger service")
}
