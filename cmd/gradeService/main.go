package main

import (
	"context"
	"distributedDemo/grades"
	"distributedDemo/logger"
	"distributedDemo/registry"
	"distributedDemo/service"
	"fmt"
	"log"
)

func main() {
	host, port := "localhost", ":5000"
	serviceAddress := fmt.Sprintf("http://%s%s", host, port)

	r := registry.Registration{
		ServiceName:      registry.GradeService,
		ServiceURL:       serviceAddress,
		RequiredServices: []registry.ServiceName{registry.LoggerService},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat",
	}
	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		grades.RegisterHandlers,
	)
	if err != nil {
		log.Fatalln("starting", registry.GradeService, ":", err)
	}
	if logProvider, err := registry.GetProvider(registry.LoggerService); err == nil {
		fmt.Println("Logger service found at:", logProvider)
		logger.SetClientLogger(logProvider, r.ServiceName)
	}
	//服务启动失败或手动终止时
	<-ctx.Done()
	fmt.Println("Shutting down grades service")
}
