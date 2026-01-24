package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/metadata"
)

// JWT header constants.
const (
	// JwtUserIdHeader HTTP header name used to pass through the JWT user id (convenience field).
	JwtUserIdHeader = "x-jwt-user-id"
	// JwtUserNameHeader HTTP header name used to pass through the JWT user name.
	JwtUserNameHeader = "x-jwt-user-name"
)

// GetClaims extracts JWT claims from context (unified API, works for HTTP or gRPC).
// It returns the full claims map, from which any field can be extracted.
// Claims are built by copying all fields from incoming gRPC metadata.
func GetClaims(ctx context.Context) jwt.MapClaims {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	claims := make(jwt.MapClaims)

	// Iterate all fields in metadata.
	for key, values := range md {
		if len(values) > 0 {
			// Use metadata key as claim key, and the first value as claim value.
			claims[key] = values[0]
		}
	}
	
	// If there are no claims, return nil.
	if len(claims) == 0 {
		return nil
	}
	return claims
}

// GetUserID extracts UserId from context (unified API, works for HTTP or gRPC).
// Convenience helper: returns UserId directly.
func GetUserID(ctx context.Context) string {
	return getFromGrpcMetadata(ctx, JwtUserIdHeader)
}

// GetUserName extracts UserName from context (unified API, works for HTTP or gRPC).
// Convenience helper: returns UserName directly.
func GetUserName(ctx context.Context) string {
	return getFromGrpcMetadata(ctx, JwtUserNameHeader)
}

func getFromGrpcMetadata(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	// Compatibility: some hops add the "gateway-" prefix.
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
