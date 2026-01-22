package bootstrap

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"github.com/addls/go-base/pkg/config"
	"github.com/addls/go-base/pkg/middleware"
)

// HttpConfig HTTP 服务基础配置（嵌入 rest.RestConf）
type HttpConfig struct {
	rest.RestConf

	// 应用配置
	App config.AppConfig `json:",optional"`
}

// RouteRegister HTTP 路由注册函数
type RouteRegister func(server *rest.Server)

// HttpOption HTTP 启动选项
type HttpOption func(*httpOptions)

type httpOptions struct {
	config        *HttpConfig // 可选：如果提供则直接使用，否则从文件加载
	middlewares   []rest.Middleware
	routeRegister RouteRegister
	beforeStart   func(*rest.Server)
	afterStart    func(*rest.Server)
}

// WithHttpConfig 直接提供 HTTP 配置（可选，如果不提供则从文件加载）
func WithHttpConfig(c *HttpConfig) HttpOption {
	return func(o *httpOptions) {
		o.config = c
	}
}

// WithHttpMiddleware 添加 HTTP 中间件
func WithHttpMiddleware(m ...rest.Middleware) HttpOption {
	return func(o *httpOptions) {
		o.middlewares = append(o.middlewares, m...)
	}
}

// WithHttpRoutes 注册 HTTP 路由
func WithHttpRoutes(register RouteRegister) HttpOption {
	return func(o *httpOptions) {
		o.routeRegister = register
	}
}

// WithHttpBeforeStart HTTP 启动前回调
func WithHttpBeforeStart(fn func(*rest.Server)) HttpOption {
	return func(o *httpOptions) {
		o.beforeStart = fn
	}
}

// WithHttpAfterStart HTTP 启动后回调
func WithHttpAfterStart(fn func(*rest.Server)) HttpOption {
	return func(o *httpOptions) {
		o.afterStart = fn
	}
}

// RunHttp 启动 HTTP 服务
// 统一入口：
//   - 配置文件路径通过命令行 flag -f 控制（默认 etc/config.yaml）
//   - 其它行为通过 HttpOption 扩展（路由、中间件、回调等）
func RunHttp(opts ...HttpOption) {
	// 解析命令行参数，获取配置文件路径
	flag.Parse()

	// 应用选项
	o := &httpOptions{
		middlewares: middleware.DefaultMiddlewares(),
	}
	for _, opt := range opts {
		opt(o)
	}

	// 加载基础配置：如果 opts 中提供了配置则直接使用，否则从文件加载
	var c HttpConfig
	if o.config != nil {
		c = *o.config
	} else {
		conf.MustLoad(config.ConfigFile(), &c)
	}

	// 创建服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 注册中间件
	for _, m := range o.middlewares {
		server.Use(m)
	}

	// 启动前回调
	if o.beforeStart != nil {
		o.beforeStart(server)
	}

	// 注册路由
	if o.routeRegister != nil {
		o.routeRegister(server)
	}

	// 打印启动信息
	logx.Infof("Starting %s %s on %s:%d [%s]",
		c.App.Name,
		c.App.Version,
		c.Host,
		c.Port,
		c.App.Env,
	)

	// 启动后回调
	if o.afterStart != nil {
		o.afterStart(server)
	}

	// 启动服务
	server.Start()
}
