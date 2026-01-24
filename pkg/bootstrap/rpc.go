package bootstrap

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/addls/go-base/pkg/config"
)

// RpcConfig base configuration for the gRPC service (embeds zrpc.RpcServerConf).
type RpcConfig struct {
	zrpc.RpcServerConf

	// Application configuration.
	App config.AppConfig `json:",optional"`
}

// ServiceRegister registers gRPC services into grpc.Server.
type ServiceRegister func(server *grpc.Server)

// RpcOption options for starting the gRPC server.
type RpcOption func(*rpcOptions)

type rpcOptions struct {
	config          *RpcConfig // Optional: if provided use directly; otherwise load from file.
	interceptors    []grpc.UnaryServerInterceptor
	serviceRegister ServiceRegister
	beforeStart     func(*zrpc.RpcServer)
	afterStart      func(*zrpc.RpcServer)
}

// WithRpcConfig provides the gRPC config directly (optional; if nil, it will be loaded from file).
func WithRpcConfig(c *RpcConfig) RpcOption {
	return func(o *rpcOptions) {
		o.config = c
	}
}

// WithRpcInterceptor adds gRPC interceptors.
func WithRpcInterceptor(interceptors ...grpc.UnaryServerInterceptor) RpcOption {
	return func(o *rpcOptions) {
		o.interceptors = append(o.interceptors, interceptors...)
	}
}

// WithRpcService registers gRPC services.
func WithRpcService(register ServiceRegister) RpcOption {
	return func(o *rpcOptions) {
		o.serviceRegister = register
	}
}

// WithRpcBeforeStart callback before the gRPC server starts.
func WithRpcBeforeStart(fn func(*zrpc.RpcServer)) RpcOption {
	return func(o *rpcOptions) {
		o.beforeStart = fn
	}
}

// WithRpcAfterStart callback after the gRPC server starts.
func WithRpcAfterStart(fn func(*zrpc.RpcServer)) RpcOption {
	return func(o *rpcOptions) {
		o.afterStart = fn
	}
}

// RunRpc starts the gRPC service.
// Unified entry:
//   - Config file path is controlled by the command-line flag -f (default: etc/config.yaml)
//   - Other behaviors can be extended via RpcOption (service registration, interceptors, callbacks, etc.)
func RunRpc(opts ...RpcOption) {
	// Parse command-line flags and get the config file path.
	flag.Parse()

	// Apply options.
	o := &rpcOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// Load base config: use the provided config if present; otherwise load from file.
	var c RpcConfig
	if o.config != nil {
		c = *o.config
	} else {
		conf.MustLoad(config.ConfigFile(), &c)
	}

	// Create gRPC server.
	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// Register services.
		if o.serviceRegister != nil {
			o.serviceRegister(grpcServer)
		}
	})

	// Register interceptors.
	for _, interceptor := range o.interceptors {
		server.AddUnaryInterceptors(interceptor)
	}

	defer server.Stop()

	// Before-start callback.
	if o.beforeStart != nil {
		o.beforeStart(server)
	}

	// Log startup information.
	logx.Infof("Starting gRPC server %s %s on %s [%s]",
		c.App.Name,
		c.App.Version,
		c.ListenOn,
		c.App.Env,
	)

	// After-start callback.
	if o.afterStart != nil {
		o.afterStart(server)
	}

	// Start serving.
	server.Start()
}
