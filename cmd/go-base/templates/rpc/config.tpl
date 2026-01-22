package config

import (
	"github.com/addls/go-base/pkg/bootstrap"
)

type Config struct {
	bootstrap.RpcConfig
	
	// 添加业务配置
	// Database DatabaseConfig `json:",optional"`
	// Redis    RedisConfig    `json:",optional"`
}

// DatabaseConfig 数据库配置
// type DatabaseConfig struct {
// 	DataSource string
// }

// RedisConfig Redis 配置
// type RedisConfig struct {
// 	Host string
// 	Pass string `json:",optional"`
// }
