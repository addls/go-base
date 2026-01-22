package bootstrap

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/gateway"

	"github.com/addls/go-base/pkg/config"
)

// GatewayConfig Gateway 服务基础配置（嵌入 gateway.GatewayConf）
type GatewayConfig struct {
	gateway.GatewayConf

	// 应用配置
	App config.AppConfig `json:",optional"`
}

// RunGateway 启动 Gateway 服务
// 统一入口：
//   - 配置文件路径通过命令行 flag -f 控制（默认 etc/config.yaml）
//   - Gateway 配置通过配置文件中的 Upstreams 定义（HTTP-to-HTTP 或 HTTP-to-gRPC）
func RunGateway() {
	// 解析命令行参数，获取配置文件路径
	flag.Parse()

	// 加载 Gateway 配置
	var c GatewayConfig
	conf.MustLoad(config.ConfigFile(), &c)

	// 创建 Gateway 服务器
	gw := gateway.MustNewServer(c.GatewayConf)
	defer gw.Stop()

	// 打印启动信息
	logx.Infof("Starting Gateway %s %s on %s:%d [%s]",
		c.App.Name,
		c.App.Version,
		c.Host,
		c.Port,
		c.App.Env,
	)

	// 启动服务
	gw.Start()
}
