package bootstrap

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/gateway"

	"github.com/addls/go-base/pkg/config"
	"github.com/addls/go-base/pkg/middleware"
)

// GatewayConfig Gateway 服务基础配置（嵌入 gateway.GatewayConf）
type GatewayConfig struct {
	gateway.GatewayConf

	// JWT 配置（可选）
	Jwt struct {
		Secret    string   `json:",optional"` // JWT 密钥
		SkipPaths []string `json:",optional"` // 跳过 JWT 验证的路径列表
	} `json:",optional"`

	// 应用配置
	App config.AppConfig `json:",optional"`
}

// GatewayOption Gateway 启动选项
type GatewayOption func(*gatewayOptions)

type gatewayOptions struct {
	config      *GatewayConfig
	beforeStart func(*gateway.Server)
	afterStart  func(*gateway.Server)
}

// WithGatewayConfig 直接提供 Gateway 配置（可选，如果不提供则从文件加载）
func WithGatewayConfig(c *GatewayConfig) GatewayOption {
	return func(o *gatewayOptions) {
		o.config = c
	}
}

// WithGatewayBeforeStart Gateway 启动前回调
func WithGatewayBeforeStart(fn func(*gateway.Server)) GatewayOption {
	return func(o *gatewayOptions) {
		o.beforeStart = fn
	}
}

// WithGatewayAfterStart Gateway 启动后回调
func WithGatewayAfterStart(fn func(*gateway.Server)) GatewayOption {
	return func(o *gatewayOptions) {
		o.afterStart = fn
	}
}

// RunGateway 启动 Gateway 服务
// 统一入口：
//   - 配置文件路径通过命令行 flag -f 控制（默认 etc/config.yaml）
//   - Gateway 配置通过 GatewayOption 扩展（配置、回调等）
//   - Upstreams 通过配置文件中的 Upstreams 字段定义（HTTP-to-HTTP 或 HTTP-to-gRPC）
func RunGateway(opts ...GatewayOption) {
	// 解析命令行参数，获取配置文件路径
	flag.Parse()

	// 应用选项
	o := &gatewayOptions{}
	for _, opt := range opts {
		opt(o)
	}

	// 加载 Gateway 配置：如果 opts 中提供了配置则直接使用，否则从文件加载
	var c GatewayConfig
	if o.config != nil {
		c = *o.config
	} else {
		conf.MustLoad(config.ConfigFile(), &c)
	}

	// 创建 Gateway 服务器
	gw := gateway.MustNewServer(c.GatewayConf)
	defer gw.Stop()

	// 注册中间件（类似 http.go 的方式）
	// 如果配置了 JWT，添加 JWT 中间件
	if c.Jwt.Secret != "" {
		jwtMw := middleware.RegisterJwtMiddleware(c.Jwt.Secret, c.Jwt.SkipPaths)
		gw.Server.Use(jwtMw)
		logx.Infof("JWT middleware configured with secret (length: %d), skip paths: %v", len(c.Jwt.Secret), c.Jwt.SkipPaths)
	}
	
	// 添加统一响应格式中间件
	gw.Server.Use(middleware.ResponseMiddleware())

	// 启动前回调（可用于注册其他中间件等）
	if o.beforeStart != nil {
		o.beforeStart(gw)
	}

	// 打印启动信息
	logx.Infof("Starting Gateway %s %s on %s:%d [%s]",
		c.App.Name,
		c.App.Version,
		c.Host,
		c.Port,
		c.App.Env,
	)

	// 启动后回调
	if o.afterStart != nil {
		o.afterStart(gw)
	}

	// 启动服务
	gw.Start()
}

