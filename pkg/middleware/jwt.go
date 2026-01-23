package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/handler"
)

const (
	// JwtClaimsHeader JWT Claims 透传的 HTTP Header 名称
	JwtClaimsHeader = "X-Jwt-Claims"
	// JwtUserIdHeader JWT UserId 透传的 HTTP Header 名称（便捷字段）
	JwtUserIdHeader = "X-Jwt-User-Id"
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
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))

	return func(next http.HandlerFunc) http.HandlerFunc {
		// 使用 authorizeMiddleware 包装 next，创建一个新的 handler
		authorizedHandler := authorizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 包装 ResponseWriter 来跟踪响应状态
			rw := &responseWriter{ResponseWriter: w}

			// 从 context 中获取 JWT claims（go-zero 会将 claims 放到 context 中）
			claims := extractJwtClaims(r.Context())
			if claims != nil {
				// 将 claims 序列化为 JSON，通过 HTTP Header 透传给后端
				if claimsJson, err := json.Marshal(claims); err == nil {
					r.Header.Set(JwtClaimsHeader, string(claimsJson))
				}

				// 提取常用的 userId 字段（如果存在）
				if userId, ok := claims["userId"].(string); ok {
					r.Header.Set(JwtUserIdHeader, userId)
				} else if userId, ok := claims["user_id"].(string); ok {
					r.Header.Set(JwtUserIdHeader, userId)
				} else if userId, ok := claims["sub"].(string); ok {
					r.Header.Set(JwtUserIdHeader, userId)
				}
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

// extractJwtClaims 从 context 中提取 JWT claims
// go-zero 的 handler.Authorize 会将 claims 放到 context 中
func extractJwtClaims(ctx context.Context) jwt.MapClaims {
	// go-zero 会将 JWT claims 放到 context 中，key 可能是 "jwtClaims" 或其他
	// 这里尝试多种可能的 key
	keys := []interface{}{
		"jwtClaims",
		"claims",
		"jwt",
	}

	for _, key := range keys {
		if val := ctx.Value(key); val != nil {
			if claims, ok := val.(jwt.MapClaims); ok {
				return claims
			}
			// 如果 val 是 map[string]interface{}，尝试转换
			if m, ok := val.(map[string]interface{}); ok {
				claims := make(jwt.MapClaims)
				for k, v := range m {
					claims[k] = v
				}
				return claims
			}
		}
	}

	// 如果 context 中没有找到，尝试从 Authorization header 中解析
	// 注意：这里只是提取，不验证（验证已经由 handler.Authorize 完成）
	return nil
}

// GetJwtClaimsFromHeader 从 HTTP Header 中获取 JWT Claims（后端服务使用）
func GetJwtClaimsFromHeader(r *http.Request) jwt.MapClaims {
	claimsJson := r.Header.Get(JwtClaimsHeader)
	if claimsJson == "" {
		return nil
	}

	var claims jwt.MapClaims
	if err := json.Unmarshal([]byte(claimsJson), &claims); err != nil {
		logx.WithContext(r.Context()).Errorf("Failed to unmarshal JWT claims from header: %v", err)
		return nil
	}

	return claims
}

// GetJwtUserIdFromHeader 从 HTTP Header 中获取 UserId（后端服务使用）
func GetJwtUserIdFromHeader(r *http.Request) string {
	return r.Header.Get(JwtUserIdHeader)
}

// GetJwtClaimsFromContext 从 context 中获取 JWT Claims（兼容 go-zero 的方式）
func GetJwtClaimsFromContext(ctx context.Context) jwt.MapClaims {
	// 尝试从 context 中获取
	if claims := extractJwtClaims(ctx); claims != nil {
		return claims
	}

	// 如果 context 中没有，尝试从 request 中获取（如果 request 在 context 中）
	if r, ok := ctx.Value("request").(*http.Request); ok {
		return GetJwtClaimsFromHeader(r)
	}

	return nil
}
