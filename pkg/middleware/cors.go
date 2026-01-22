package middleware

import (
	"net/http"
)

// CorsConfig CORS 配置
type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCorsConfig 默认 CORS 配置
func DefaultCorsConfig() CorsConfig {
	return CorsConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Trace-Id"},
		ExposeHeaders:    []string{"Content-Length", "X-Trace-Id"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// Cors 跨域中间件
func Cors() func(http.HandlerFunc) http.HandlerFunc {
	return CorsWithConfig(DefaultCorsConfig())
}

// CorsWithConfig 带配置的跨域中间件
func CorsWithConfig(cfg CorsConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next(w, r)
				return
			}

			// 检查是否允许该源
			allowed := false
			for _, o := range cfg.AllowOrigins {
				if o == "*" || o == origin {
					allowed = true
					if o == "*" {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					} else {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}
					break
				}
			}

			if !allowed {
				next(w, r)
				return
			}

			w.Header().Set("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods))
			w.Header().Set("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders))
			w.Header().Set("Access-Control-Expose-Headers", joinStrings(cfg.ExposeHeaders))

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// 预检请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next(w, r)
		}
	}
}

func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}
