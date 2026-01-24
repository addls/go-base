package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"github.com/addls/go-base/pkg/errcode"
	"github.com/addls/go-base/pkg/response"
)

// responseWrapper intercepts responses for unified format conversion.
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	written    bool
}

func newResponseWrapper(w http.ResponseWriter) *responseWrapper {
	return &responseWrapper{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		body:           &bytes.Buffer{},
	}
}

func (w *responseWrapper) WriteHeader(code int) {
	if !w.written {
		w.statusCode = code
		w.written = true
	}
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	// Write to buffer instead of writing to downstream directly.
	return w.body.Write(b)
}

// ResponseMiddleware is a unified response format middleware.
// It wraps backend responses into the unified response.Response format.
func ResponseMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rw := newResponseWrapper(w)

			// Execute downstream handler (forward to backend service).
			next(rw, r)

			// Read original response.
			rawBody := rw.body.Bytes()
			status := rw.statusCode

			// If the original response is empty, return a unified empty success response.
			if len(rawBody) == 0 {
				response.Ok(w)
				return
			}

			// Check whether it's already in unified format (contains the "code" field).
			var maybeUnified struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(rawBody, &maybeUnified); err == nil {
				// If it already has the code field, treat it as unified format and pass through.
				if maybeUnified.Code != 0 || len(rawBody) > 0 {
					// Check whether it contains the msg field (a stricter check).
					var checkMsg struct {
						Code int    `json:"code"`
						Msg  string `json:"msg"`
					}
					if err := json.Unmarshal(rawBody, &checkMsg); err == nil && checkMsg.Msg != "" {
						// Already unified format; pass through directly.
						w.WriteHeader(status)
						_, _ = w.Write(rawBody)
						return
					}
				}
			}

			// Determine success/failure by HTTP status code.
			if status >= http.StatusBadRequest {
				// Error response: try to parse error information.
				var errData interface{}
				if err := json.Unmarshal(rawBody, &errData); err != nil {
					errData = string(rawBody)
				}

				// Try to extract an error message.
				msg := http.StatusText(status)
				if m, ok := errData.(map[string]interface{}); ok {
					if emsg, ok := m["message"].(string); ok && emsg != "" {
						msg = emsg
					} else if emsg, ok := m["error"].(string); ok && emsg != "" {
						msg = emsg
					}
				}

				// Map HTTP status code to business error code.
				var code int
				switch status {
				case http.StatusBadRequest:
					code = errcode.ErrInvalidParam.Code
				case http.StatusUnauthorized:
					code = errcode.ErrUnauthorized.Code
				case http.StatusForbidden:
					code = errcode.ErrForbidden.Code
				case http.StatusNotFound:
					code = errcode.ErrNotFound.Code
				case http.StatusInternalServerError:
					code = errcode.ErrInternal.Code
				default:
					code = errcode.ErrInternal.Code
				}

				response.ErrorWithCode(w, code, msg)
				return
			}

			// Success response: wrap the original payload into the data field.
			var data interface{}
			if err := json.Unmarshal(rawBody, &data); err != nil {
				// If it's not JSON, return it as a string.
				data = string(rawBody)
			}

			// Return unified success response.
			response.OkWithData(w, data)
		}
	}
}
