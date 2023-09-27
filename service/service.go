package service

import (
	"context"
	"distributedDemo/registry"
	"fmt"
	"log"
	"net/http"
)

// Start 启动多个webserver服务
func Start(ctx context.Context, host, port string,
	reg registry.Registration, registerHandlersFunc func()) (context.Context, error) {
	registerHandlersFunc()
	ctx = startService(ctx, reg.ServiceName, host, port)
	err := registry.RegisterService(reg)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

// 操作服务状态（启动或关闭）(包级函数）
// 这段代码实现了启动一个 HTTP 服务器的功能。它使用了 context.Context 来处理关闭服务器的操作。代码的执行流程如下：
// 首先，通过 context.WithCancel 创建一个新的 context.Context，并启动一个新的 goroutine。
// 在新的 goroutine 中，调用 srv.ListenAndServe 方法来启动 HTTP 服务器。如果服务器无法启动，则调用 registry.ShutdownService 方法来关闭服务器，并取消上下文。
// 然后，另外启动一个 goroutine，等待用户输入。如果用户按下任意键，则调用 registry.ShutdownService 方法关闭服务器，并调用 srv.Shutdown 方法取消上下文。
// 需要注意的是，在这段代码中，srv.Addr 变量并未使用到 host 变量，也就是说，只有 port 参数被用来作为服务器的端口。
func startService(ctx context.Context, serviceName registry.ServiceName,
	host, port string) context.Context {

	ctx, cancel := context.WithCancel(ctx)
	var srv http.Server
	//host+port
	srv.Addr = port

	go func() {
		log.Println(srv.ListenAndServe())
		err := registry.ShutdownService(fmt.Sprintf("http://%s%s", host, port))
		if err != nil {
			log.Println("func startService:", err)
		}
		cancel()
	}()

	go func() {
		fmt.Printf("%v started.Press any key to stop...\n", serviceName)
		var s string
		fmt.Scanln(&s)
		err := registry.ShutdownService(fmt.Sprintf("http://%s%s", host, port))
		if err != nil {
			log.Println("func startService:", err)
		}
		_ = srv.Shutdown(ctx)
		cancel()
	}()

	return ctx
}
