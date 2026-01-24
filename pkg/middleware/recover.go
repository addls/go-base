package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/addls/go-base/pkg/errcode"
	"github.com/addls/go-base/pkg/response"
)

// Recover is a panic recovery middleware.
func Recover() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logx.WithContext(r.Context()).Errorf("panic recovered: %v\n%s", err, debug.Stack())
					response.Error(w, errcode.ErrInternal)
				}
			}()
			next(w, r)
		}
	}
}
