package main

import (
	"github.com/addls/go-base/pkg/bootstrap"
	"google.golang.org/grpc"
	{{.imports}}
)

func main() {
	// 使用 go-base 统一的启动入口：
	// 配置文件 flag：-f 默认路径：etc/config.yaml
	bootstrap.RunRpc(bootstrap.WithRpcService(func(grpcServer *grpc.Server) {
		server.RegisterServices(grpcServer, svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]()))
	}))
}
