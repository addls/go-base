package bootstrap

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/gateway"

	"github.com/addls/go-base/pkg/config"
	"github.com/addls/go-base/pkg/middleware"
)

// GatewayConfig base configuration for the Gateway service (embeds gateway.GatewayConf).
type GatewayConfig struct {
	gateway.GatewayConf

	// JWT configuration (optional).
	Jwt struct {
		Secret    string   `json:",optional"` // JWT secret
		SkipPaths []string `json:",optional"` // Paths that skip JWT verification
	} `json:",optional"`

	// Application configuration.
	App config.AppConfig `json:",optional"`
}

// GatewayOption options for starting the Gateway.
type GatewayOption func(*gatewayOptions)

type gatewayOptions struct {
	config      *GatewayConfig
	beforeStart func(*gateway.Server)
	afterStart  func(*gateway.Server)
}

// WithGatewayConfig provides the Gateway config directly (optional; if nil, it will be loaded from file).
func WithGatewayConfig(c *GatewayConfig) GatewayOption {
	return func(o *gatewayOptions) {
		o.config = c
	}
}

// WithGatewayBeforeStart callback before the Gateway starts.
func WithGatewayBeforeStart(fn func(*gateway.Server)) GatewayOption {
	return func(o *gatewayOptions) {
		o.beforeStart = fn
	}
}

// WithGatewayAfterStart callback after the Gateway starts.
func WithGatewayAfterStart(fn func(*gateway.Server)) GatewayOption {
	return func(o *gatewayOptions) {
		o.afterStart = fn
	}
}

// RunGateway starts the Gateway service.
// Unified entry:
//   - Config file path is controlled by the command-line flag -f (default: etc/config.yaml)
//   - Gateway behavior can be extended via GatewayOption (config, callbacks, etc.)
//   - Upstreams are defined by the Upstreams field in config (HTTP-to-HTTP or HTTP-to-gRPC)
func RunGateway(opts ...GatewayOption) {
	// Parse command-line flags and get the config file path.
	flag.Parse()

	// Apply options.
	o := &gatewayOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Load Gateway config: use the provided config if present; otherwise load from file.
	var c GatewayConfig
	if o.config != nil {
		c = *o.config
	} else {
		conf.MustLoad(config.ConfigFile(), &c)
	}

	// Create the Gateway server.
	gw := gateway.MustNewServer(c.GatewayConf)
	defer gw.Stop()

	// Register middlewares (similar to http.go).
	// If JWT is configured, add the JWT middleware.
	if c.Jwt.Secret != "" {
		jwtMw := middleware.RegisterJwtMiddleware(c.Jwt.Secret, c.Jwt.SkipPaths)
		gw.Server.Use(jwtMw)
		logx.Infof("JWT middleware configured with secret (length: %d), skip paths: %v", len(c.Jwt.Secret), c.Jwt.SkipPaths)
	}
	
	// Add unified response format middleware.
	gw.Server.Use(middleware.ResponseMiddleware())

	// Before-start callback (can be used to register other middlewares, etc.).
	if o.beforeStart != nil {
		o.beforeStart(gw)
	}

	// Log startup information.
	logx.Infof("Starting Gateway %s %s on %s:%d [%s]",
		c.App.Name,
		c.App.Version,
		c.Host,
		c.Port,
		c.App.Env,
	)

	// After-start callback.
	if o.afterStart != nil {
		o.afterStart(gw)
	}

	// Start serving.
	gw.Start()
}

