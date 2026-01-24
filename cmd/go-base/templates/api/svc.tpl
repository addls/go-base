package svc

import (
	{{.configImport}}
)

type ServiceContext struct {
	Config *config.Config
	// Add business dependencies.
	// UserModel model.UserModel
	// Redis     *redis.Redis
}

func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		// Initialize business dependencies.
	}
}
