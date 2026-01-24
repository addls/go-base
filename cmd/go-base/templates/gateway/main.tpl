package main

import (
	"github.com/addls/go-base/pkg/bootstrap"
)

func main() {
	// Use the unified go-base startup entry:
	// Config file flag: -f, default path: etc/config.yaml
	bootstrap.RunGateway()
}
