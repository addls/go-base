package main

import (
	{{.importPackages}}
	"github.com/addls/go-base/pkg/bootstrap"
)

func main() {
	// Use the unified go-base startup entry:
	// Config file flag: -f, default path: etc/config.yaml
	bootstrap.RunHttp(bootstrap.WithHttpRoutes(func(server *rest.Server) {
		handler.RegisterHandlers(server, svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]()))
	}))
}
