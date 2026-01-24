// Package config provides shared utilities for configuration.
package config

import (
	"flag"
)

// Unified config file flag definition, defaulting to the go-zero convention: etc/config.yaml.
var configFile = flag.String("f", "etc/config.yaml", "the config file")

// AppConfig application configuration.
type AppConfig struct {
	Name    string `json:",default=app"`
	Version string `json:",default=1.0.0"`
	Env     string `json:",default=dev"` // dev, test, prod
}

// ConfigFile returns the current config file path (from -f flag or default).
func ConfigFile() string {
	return *configFile
}
