package config

import (
	"github.com/addls/go-base/pkg/bootstrap"
)

type Config struct {
	bootstrap.HttpConfig
	
	// Add business configuration.
	// Database DatabaseConfig `json:",optional"`
	// Redis    RedisConfig    `json:",optional"`
}

// DatabaseConfig database configuration.
// type DatabaseConfig struct {
// 	DataSource string
// }

// RedisConfig Redis configuration.
// type RedisConfig struct {
// 	Host string
// 	Pass string `json:",optional"`
// }
