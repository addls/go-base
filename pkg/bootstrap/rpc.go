package bootstrap

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"

	"github.com/addls/go-base/pkg/config"
)

// RpcConfig gRPC 服务基础配置（嵌入 zrpc.RpcServerConf）
type RpcConfig struct {
	zrpc.RpcServerConf

	// 应用配置
	App config.AppConfig `json:",optional"`
}

// ServiceRegister gRPC 服务注册函数（注册到 grpc.Server）
type ServiceRegister func(server *grpc.Server)

// RpcOption gRPC 启动选项
type RpcOption func(*rpcOptions)

type rpcOptions struct {
	config          *RpcConfig // 可选：如果提供则直接使用，否则从文件加载
	interceptors    []grpc.UnaryServerInterceptor
	serviceRegister ServiceRegister
	beforeStart     func(*zrpc.RpcServer)
	afterStart      func(*zrpc.RpcServer)
}

// WithRpcConfig 直接提供 gRPC 配置（可选，如果不提供则从文件加载）
func WithRpcConfig(c *RpcConfig) RpcOption {
	return func(o *rpcOptions) {
		o.config = c
	}
}

// WithRpcInterceptor 添加 gRPC 拦截器
func WithRpcInterceptor(interceptors ...grpc.UnaryServerInterceptor) RpcOption {
	return func(o *rpcOptions) {
		o.interceptors = append(o.interceptors, interceptors...)
	}
}

// WithRpcService 注册 gRPC 服务
func WithRpcService(register ServiceRegister) RpcOption {
	return func(o *rpcOptions) {
		o.serviceRegister = register
	}
}

// WithRpcBeforeStart gRPC 启动前回调
func WithRpcBeforeStart(fn func(*zrpc.RpcServer)) RpcOption {
	return func(o *rpcOptions) {
		o.beforeStart = fn
	}
}

// WithRpcAfterStart gRPC 启动后回调
func WithRpcAfterStart(fn func(*zrpc.RpcServer)) RpcOption {
	return func(o *rpcOptions) {
		o.afterStart = fn
	}
}

// RunRpc 启动 gRPC 服务
// 统一入口：
//   - 配置文件路径通过命令行 flag -f 控制（默认 etc/config.yaml）
//   - 其它行为通过 RpcOption 扩展（服务注册、拦截器、回调等）
func RunRpc(opts ...RpcOption) {
	// 解析命令行参数，获取配置文件路径
	flag.Parse()

	// 应用选项
	o := &rpcOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// 加载基础配置：如果 opts 中提供了配置则直接使用，否则从文件加载
	var c RpcConfig
	if o.config != nil {
		c = *o.config
	} else {
		conf.MustLoad(config.ConfigFile(), &c)
	}

	// 创建 gRPC 服务器
	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 注册服务
		if o.serviceRegister != nil {
			o.serviceRegister(grpcServer)
		}
	})

	// 注册拦截器
	for _, interceptor := range o.interceptors {
		server.AddUnaryInterceptors(interceptor)
	}

	defer server.Stop()

	// 启动前回调
	if o.beforeStart != nil {
		o.beforeStart(server)
	}

	// 打印启动信息
	logx.Infof("Starting gRPC server %s %s on %s [%s]",
		c.App.Name,
		c.App.Version,
		c.ListenOn,
		c.App.Env,
	)

	// 启动后回调
	if o.afterStart != nil {
		o.afterStart(server)
	}

	// 启动服务
	server.Start()
}
