package config

import (
	"github.com/addls/go-base/pkg/bootstrap"
	{{.authImport}}
)

type Config struct {
	bootstrap.HttpConfig
	{{.auth}}
}
