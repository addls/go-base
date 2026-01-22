package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/addls/go-base/pkg/errcode"
	"github.com/addls/go-base/pkg/response"
)

type contextKey string

const (
	UserIDKey   contextKey = "userId"
	UserNameKey contextKey = "userName"
	UserRoleKey contextKey = "userRole"
)

// TokenValidator Token 验证器接口
type TokenValidator interface {
	Validate(token string) (*UserInfo, error)
}

// UserInfo 用户信息
type UserInfo struct {
	UserID   string
	UserName string
	Role     string
}

// Auth 认证中间件
func Auth(validator TokenValidator) func(http.HandlerFunc) http.HandlerFunc {
	return AuthWithConfig(validator, AuthConfig{})
}

// AuthConfig 认证配置
type AuthConfig struct {
	SkipPaths   []string // 跳过的路径
	TokenHeader string   // Token 请求头
	TokenPrefix string   // Token 前缀
}

// AuthWithConfig 带配置的认证中间件
func AuthWithConfig(validator TokenValidator, cfg AuthConfig) func(http.HandlerFunc) http.HandlerFunc {
	if cfg.TokenHeader == "" {
		cfg.TokenHeader = "Authorization"
	}
	if cfg.TokenPrefix == "" {
		cfg.TokenPrefix = "Bearer"
	}

	skipMap := make(map[string]bool)
	for _, path := range cfg.SkipPaths {
		skipMap[path] = true
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 跳过的路径
			if skipMap[r.URL.Path] {
				next(w, r)
				return
			}

			// 获取 Token
			authHeader := r.Header.Get(cfg.TokenHeader)
			if authHeader == "" {
				response.Error(w, errcode.ErrTokenMissing)
				return
			}

			// 解析 Token
			token := authHeader
			if cfg.TokenPrefix != "" {
				if !strings.HasPrefix(authHeader, cfg.TokenPrefix+" ") {
					response.Error(w, errcode.ErrTokenInvalid)
					return
				}
				token = strings.TrimPrefix(authHeader, cfg.TokenPrefix+" ")
			}

			// 验证 Token
			userInfo, err := validator.Validate(token)
			if err != nil {
				response.Error(w, errcode.FromError(err))
				return
			}

			// 注入用户信息到 Context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userInfo.UserID)
			ctx = context.WithValue(ctx, UserNameKey, userInfo.UserName)
			ctx = context.WithValue(ctx, UserRoleKey, userInfo.Role)

			next(w, r.WithContext(ctx))
		}
	}
}

// GetUserID 从 Context 获取用户 ID
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserName 从 Context 获取用户名
func GetUserName(ctx context.Context) string {
	if name, ok := ctx.Value(UserNameKey).(string); ok {
		return name
	}
	return ""
}

// GetUserRole 从 Context 获取用户角色
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}
