// Package config 提供配置相关的公共功能
package config

import (
	"flag"
)

// 统一配置文件 flag 定义，默认使用 go-zero 常见约定：etc/config.yaml
var configFile = flag.String("f", "etc/config.yaml", "the config file")

// AppConfig 应用配置
type AppConfig struct {
	Name    string `json:",default=app"`
	Version string `json:",default=1.0.0"`
	Env     string `json:",default=dev"` // dev, test, prod
}

// ConfigFile 返回当前配置文件路径（由命令行 -f 或默认值决定）
func ConfigFile() string {
	return *configFile
}
