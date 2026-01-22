package bootstrap

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"github.com/addls/go-base/pkg/config"
)

// ServiceContext 服务上下文基础结构
// 业务项目可以嵌入此结构
type ServiceContext struct {
	Config interface{}
}

// NewServiceContext 创建服务上下文
func NewServiceContext(config interface{}) *ServiceContext {
	return &ServiceContext{
		Config: config,
	}
}

// RegisterRoutes 路由注册辅助函数
func RegisterRoutes(server *rest.Server, routes []rest.Route) {
	server.AddRoutes(routes)
}

// RegisterRoutesWithPrefix 带前缀的路由注册
func RegisterRoutesWithPrefix(server *rest.Server, prefix string, routes []rest.Route) {
	server.AddRoutes(routes, rest.WithPrefix(prefix))
}

// RegisterRoutesWithJwt 带 JWT 验证的路由注册
func RegisterRoutesWithJwt(server *rest.Server, routes []rest.Route, jwtSecret string) {
	server.AddRoutes(routes, rest.WithJwt(jwtSecret))
}

// MustLoadConfig 泛型封装：从统一配置文件加载任意配置结构体
// 配置文件路径由 -f flag 控制（默认 etc/config.yaml）
func MustLoadConfig[T any]() *T {
	var c T
	conf.MustLoad(config.ConfigFile(), &c)
	return &c
}
