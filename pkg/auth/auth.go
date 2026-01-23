package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
)

// JWT Header 常量
const (
	// JwtUserIdHeader JWT UserId 透传的 HTTP Header 名称（便捷字段）
	JwtUserIdHeader = "x-jwt-user-id"
	// JwtUserNameHeader JWT UserName 透传的 HTTP Header 名称
	JwtUserNameHeader = "x-jwt-user-name"
)

// GetClaims 从 context 中获取 JWT Claims（统一接口，自动识别 HTTP 或 gRPC）
// 返回完整的 Claims，可以从中提取任何字段
// 直接从 metadata 中提取所有字段构建 claims
func GetClaims(ctx context.Context) jwt.MapClaims {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	claims := make(jwt.MapClaims)

	// 直接遍历 metadata 中的所有字段
	for key, values := range md {
		if len(values) > 0 {
			// 直接使用 key 作为 claims 的 key，value 作为 claims 的 value
			claims[key] = values[0]
		}
	}
	
	// 如果没有任何 claims，返回 nil
	if len(claims) == 0 {
		return nil
	}
	return claims
}

// GetUserID 从 context 中获取 UserId（统一接口，自动识别 HTTP 或 gRPC）
// 便捷方法，直接返回 UserId
func GetUserID(ctx context.Context) string {
	return getFromGrpcMetadata(ctx, JwtUserIdHeader)
}

// GetUserName 从 context 中获取 UserName（统一接口，自动识别 HTTP 或 gRPC）
// 便捷方法，直接返回 UserName
func GetUserName(ctx context.Context) string {
	return getFromGrpcMetadata(ctx, JwtUserNameHeader)
}

func getFromGrpcMetadata(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// 兼容：有的链路会加 "gateway-" 前缀
	keys := []string{
		key,            // x-jwt-...
		"gateway-" + key, // gateway-x-jwt-...
	}
	for _, k := range keys {
		if values := md.Get(k); len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
