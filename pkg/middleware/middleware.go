// Package middleware 提供统一的 HTTP 中间件
package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterGlobalMiddleware 注册全局中间件到 go-zero server
func RegisterGlobalMiddleware(server *rest.Server, middlewares ...rest.Middleware) {
	for _, m := range middlewares {
		server.Use(m)
	}
}

// DefaultMiddlewares 默认中间件列表
func DefaultMiddlewares() []rest.Middleware {
	return []rest.Middleware{
		RecoverMiddleware,
		CorsMiddleware,
	}
}

// RecoverMiddleware 恢复中间件
func RecoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return Recover()(next)
}

// CorsMiddleware CORS 中间件
func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return Cors()(next)
}
