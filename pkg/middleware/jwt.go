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

// JwtConfig JWT configuration.
type JwtConfig struct {
	Secret    string   // JWT secret
	SkipPaths []string // Paths that skip JWT verification
}

// responseWriter wraps http.ResponseWriter to track whether a response has been written.
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

// RegisterJwtMiddleware registers a JWT middleware for the Gateway.
// It uses go-zero's handler.Authorize to verify JWT.
// After successful verification, JWT claims are passed through to backend services via HTTP headers.
func RegisterJwtMiddleware(secret string, skipPaths []string) rest.Middleware {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}

	// Use go-zero's Authorize handler (returns a middleware function).
	authorizeMiddleware := handler.Authorize(secret, handler.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, err error) {
		logx.WithContext(r.Context()).Errorf("JWT authorization failed: %v", err)
		// Use the unified error response format.
		response.ErrorWithCode(w, errcode.ErrUnauthorized.Code, errcode.ErrUnauthorized.Msg)
	}))

	return func(next http.HandlerFunc) http.HandlerFunc {
		// Wrap next with authorizeMiddleware to create a new handler.
		authorizedHandler := authorizeMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap ResponseWriter to track response status.
			rw := &responseWriter{ResponseWriter: w}

			// Extract JWT claims from context (non-standard fields).
			// go-zero's handler.Authorize stores each non-standard field into context using the field name as the key.
			// Standard fields (sub, exp, iat, iss, aud, nbf, jti) are ignored.
			ctx := r.Context()
			
			// Extract user info (using non-standard fields uid and name),
			// and pass through to backend services via HTTP headers.
			if uid, ok := ctx.Value("uid").(string); ok && uid != "" {
				// grpc-gateway forwarding into gRPC metadata requires the "Grpc-Metadata-" prefix.
				r.Header.Set("Grpc-Metadata-"+auth.JwtUserIdHeader, uid)
			}
			if name, ok := ctx.Value("name").(string); ok && name != "" {
				// grpc-gateway forwarding into gRPC metadata requires the "Grpc-Metadata-" prefix.
				r.Header.Set("Grpc-Metadata-"+auth.JwtUserNameHeader, name)
			}

			// Continue to the next handler.
			next(rw, r)
		}))

		return func(w http.ResponseWriter, r *http.Request) {
			// Skip paths.
			if skipMap[r.URL.Path] {
				next(w, r)
				return
			}

			// Run authorization and the subsequent handler chain.
			authorizedHandler.ServeHTTP(w, r)
		}
	}
}


