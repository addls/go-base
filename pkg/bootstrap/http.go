package bootstrap

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"github.com/addls/go-base/pkg/config"
	"github.com/addls/go-base/pkg/middleware"
)

// HttpConfig base configuration for the HTTP service (embeds rest.RestConf).
type HttpConfig struct {
	rest.RestConf

	// Application configuration.
	App config.AppConfig `json:",optional"`
}

// RouteRegister registers HTTP routes.
type RouteRegister func(server *rest.Server)

// HttpOption options for starting the HTTP server.
type HttpOption func(*httpOptions)

type httpOptions struct {
	config        *HttpConfig // Optional: if provided use directly; otherwise load from file.
	middlewares   []rest.Middleware
	routeRegister RouteRegister
	beforeStart   func(*rest.Server)
	afterStart    func(*rest.Server)
}

// WithHttpConfig provides the HTTP config directly (optional; if nil, it will be loaded from file).
func WithHttpConfig(c *HttpConfig) HttpOption {
	return func(o *httpOptions) {
		o.config = c
	}
}

// WithHttpMiddleware adds HTTP middlewares.
func WithHttpMiddleware(m ...rest.Middleware) HttpOption {
	return func(o *httpOptions) {
		o.middlewares = append(o.middlewares, m...)
	}
}

// WithHttpRoutes registers HTTP routes.
func WithHttpRoutes(register RouteRegister) HttpOption {
	return func(o *httpOptions) {
		o.routeRegister = register
	}
}

// WithHttpBeforeStart callback before the HTTP server starts.
func WithHttpBeforeStart(fn func(*rest.Server)) HttpOption {
	return func(o *httpOptions) {
		o.beforeStart = fn
	}
}

// WithHttpAfterStart callback after the HTTP server starts.
func WithHttpAfterStart(fn func(*rest.Server)) HttpOption {
	return func(o *httpOptions) {
		o.afterStart = fn
	}
}

// RunHttp starts the HTTP service.
// Unified entry:
//   - Config file path is controlled by the command-line flag -f (default: etc/config.yaml)
//   - Other behaviors can be extended via HttpOption (routes, middlewares, callbacks, etc.)
func RunHttp(opts ...HttpOption) {
	// Parse command-line flags and get the config file path.
	flag.Parse()

	// Apply options.
	o := &httpOptions{
		middlewares: middleware.DefaultMiddlewares(),
	}
	for _, opt := range opts {
		opt(o)
	}

	// Load base config: use the provided config if present; otherwise load from file.
	var c HttpConfig
	if o.config != nil {
		c = *o.config
	} else {
		conf.MustLoad(config.ConfigFile(), &c)
	}

	// Create server.
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// Register middlewares.
	for _, m := range o.middlewares {
		server.Use(m)
	}

	// Before-start callback.
	if o.beforeStart != nil {
		o.beforeStart(server)
	}

	// Register routes.
	if o.routeRegister != nil {
		o.routeRegister(server)
	}

	// Log startup information.
	logx.Infof("Starting %s %s on %s:%d [%s]",
		c.App.Name,
		c.App.Version,
		c.Host,
		c.Port,
		c.App.Env,
	)

	// After-start callback.
	if o.afterStart != nil {
		o.afterStart(server)
	}

	// Start serving.
	server.Start()
}
