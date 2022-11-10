package main

import (
	"context"
	"distributedDemo/grade"
	"distributedDemo/registry"
	"distributedDemo/service"
	"fmt"
	"log"
)

func main() {
	host, port := "localhost", ":5000"
	serviceAddress := fmt.Sprintf("http://%s%s", host, port)

	r := registry.Registration{
		ServiceName: registry.GradeService,
		ServiceURL:  serviceAddress,
	}
	ctx, err := service.Start(
		context.Background(),
		host,
		port,
		r,
		grade.RegisterHandlers,
	)
	if err != nil {
		log.Fatalln("starting", registry.GradeService, ":", err)
	}
	//服务启动失败或手动终止时
	<-ctx.Done()
	fmt.Println("Shutting down grade service")
}
