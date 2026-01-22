package svc

import (
	{{.configImport}}
)

type ServiceContext struct {
	Config *config.Config
	// 添加业务依赖
	// UserModel model.UserModel
	// Redis     *redis.Redis
}

func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		// 初始化业务依赖
	}
}
