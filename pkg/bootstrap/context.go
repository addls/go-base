package bootstrap

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"github.com/addls/go-base/pkg/config"
)

// ServiceContext is the base structure for service context.
// Business projects can embed this struct.
type ServiceContext struct {
	Config interface{}
}

// NewServiceContext creates a service context.
func NewServiceContext(config interface{}) *ServiceContext {
	return &ServiceContext{
		Config: config,
	}
}

// RegisterRoutes is a helper to register routes.
func RegisterRoutes(server *rest.Server, routes []rest.Route) {
	server.AddRoutes(routes)
}

// RegisterRoutesWithPrefix registers routes with a prefix.
func RegisterRoutesWithPrefix(server *rest.Server, prefix string, routes []rest.Route) {
	server.AddRoutes(routes, rest.WithPrefix(prefix))
}

// RegisterRoutesWithJwt registers routes with JWT verification.
func RegisterRoutesWithJwt(server *rest.Server, routes []rest.Route, jwtSecret string) {
	server.AddRoutes(routes, rest.WithJwt(jwtSecret))
}

// MustLoadConfig is a generic helper that loads any config struct from the unified config file.
// The config file path is controlled by the -f flag (default: etc/config.yaml).
func MustLoadConfig[T any]() *T {
	var c T
	conf.MustLoad(config.ConfigFile(), &c)
	return &c
}
