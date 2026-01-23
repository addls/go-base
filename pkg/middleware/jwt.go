package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/handler"

	"github.com/addls/go-base/pkg/auth"
	"github.com/addls/go-base/pkg/errcode"
	"github.com/addls/go-base/pkg/response"
)

// JwtConfig JWT 配置
type JwtConfig struct {
	Secret    string   // JWT 密钥
	SkipPaths []string // 跳过 JWT 验证的路径列表
}

// responseWriter 包装 http.ResponseWriter 来跟踪是否已写入响应
type responseWriter struct {
	http.ResponseWriter
	written bool
	status  int
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.status = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
		rw.status = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// RegisterJwtMiddleware 注册 JWT 中间件到 Gateway
// 使用 go-zero 的 handler.Authorize 进行 JWT 验证
// 验证成功后，将 JWT claims 通过 HTTP Header 透传给后端服务
func RegisterJwtMiddleware(secret string, skipPaths []string) rest.Middleware {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	// 使用 go-zero 的 Authorize handler（返回中间件函数）
	authorizeMiddleware := handler.Authorize(secret, handler.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, err error) {
		logx.WithContext(r.Context()).Errorf("JWT authorization failed: %v", err)
		// 使用统一的错误响应格式
		response.ErrorWithCode(w, errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg)
	}))

	return func(next http.HandlerFunc) http.HandlerFunc {
		// 使用 authorizeMiddleware 包装 next，创建一个新的 handler
		authorizedHandler := authorizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 包装 ResponseWriter 来跟踪响应状态
			rw := &responseWriter{ResponseWriter: w}

			// 从 context 中提取 JWT claims（非标准字段）
			// go-zero 的 handler.Authorize 会将非标准字段逐个放到 context 中，key 就是字段名
			// 标准字段（sub, exp, iat, iss, aud, nbf, jti）会被忽略
			ctx := r.Context()
			
			// 提取用户信息（使用非标准字段 uid 和 name）
			// 通过 HTTP Header 透传给后端服务
			if uid, ok := ctx.Value("uid").(string); ok && uid != "" {
				// grpc-gateway 转发到 gRPC metadata 需要 Grpc-Metadata- 前缀
				r.Header.Set("Grpc-Metadata-"+auth.JwtUserIdHeader, uid)
			}
			if name, ok := ctx.Value("name").(string); ok && name != "" {
				// grpc-gateway 转发到 gRPC metadata 需要 Grpc-Metadata- 前缀
				r.Header.Set("Grpc-Metadata-"+auth.JwtUserNameHeader, name)
			}

			// 继续执行下一个 handler
			next(rw, r)
		}))

		return func(w http.ResponseWriter, r *http.Request) {
			// 跳过的路径
			if skipMap[r.URL.Path] {
				next(w, r)
				return
			}

			// 执行授权验证和后续处理
			authorizedHandler.ServeHTTP(w, r)
		}
	}
}


