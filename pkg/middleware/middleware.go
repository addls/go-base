// Package middleware provides unified HTTP middlewares.
package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterGlobalMiddleware registers global middlewares into the go-zero server.
func RegisterGlobalMiddleware(server *rest.Server, middlewares ...rest.Middleware) {
	for _, m := range middlewares {
		server.Use(m)
	}
}

// DefaultMiddlewares returns the default middleware list.
func DefaultMiddlewares() []rest.Middleware {
	return []rest.Middleware{
		RecoverMiddleware,
		CorsMiddleware,
	}
}

// RecoverMiddleware is a panic recovery middleware.
func RecoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return Recover()(next)
}

// CorsMiddleware is a CORS middleware.
func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return Cors()(next)
}
