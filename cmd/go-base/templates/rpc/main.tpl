package main

import (
	"github.com/addls/go-base/pkg/bootstrap"
	"google.golang.org/grpc"
	{{.imports}}
)

func main() {
	// Use the unified go-base startup entry:
	// Config file flag: -f, default path: etc/config.yaml
	bootstrap.RunRpc(bootstrap.WithRpcService(func(grpcServer *grpc.Server) {
		server.RegisterServices(grpcServer, svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]()))
	}))
}
