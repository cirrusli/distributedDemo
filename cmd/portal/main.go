package main

import (
	"context"
	"distributedDemo/logger"
	"distributedDemo/portal"
	"distributedDemo/registry"
	"distributedDemo/service"
	"fmt"
	"log"
)

func main() {
	err := portal.ImportTemplates()
	if err != nil {
		log.Fatalln("In ./cmd/portal: func main:", err)
	}
	host, port := "localhost", ":6000"
	serviceAddress := fmt.Sprintf("http://%s%s", host, port)

	r := registry.Registration{
		ServiceName: registry.PortalService,
		ServiceURL:  serviceAddress,
		RequiredServices: []registry.ServiceName{
			registry.LoggerService,
			registry.GradeService,
		},
		ServiceUpdateURL: serviceAddress + "/services",
		HeartbeatURL:     serviceAddress + "/heartbeat",
	}

	ctx, err := service.Start(context.Background(),
		host,
		port,
		r,
		portal.RegisterHandlers)
	if err != nil {
		log.Fatalln("In ./cmd/portal: func main:", err)
	}
	if logProvider, err := registry.GetProvider(registry.LoggerService); err != nil {
		logger.SetClientLogger(logProvider, r.ServiceName)
	}
	<-ctx.Done()
	fmt.Println("Shutting down portal...")
}
