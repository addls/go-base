package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"github.com/addls/go-base/pkg/errcode"
	"github.com/addls/go-base/pkg/response"
)

// responseWrapper 拦截响应，用于统一格式转换
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
	// 写入缓冲区，不直接写到下游
	return w.body.Write(b)
}

// ResponseMiddleware 统一响应格式中间件
// 将 RPC 返回的业务数据包装成统一的 response.Response 格式
func ResponseMiddleware() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rw := newResponseWrapper(w)

			// 执行下游 handler（转发到后端服务）
			next(rw, r)

			// 读取原始响应
			rawBody := rw.body.Bytes()
			status := rw.statusCode

			// 如果原始响应为空，返回统一的空成功响应
			if len(rawBody) == 0 {
				response.Ok(w)
				return
			}

			// 检查是否已经是统一格式（包含 code 字段）
			var maybeUnified struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(rawBody, &maybeUnified); err == nil {
				// 如果已经有 code 字段，说明已经是统一格式，直接透传
				if maybeUnified.Code != 0 || len(rawBody) > 0 {
					// 检查是否包含 msg 字段（更严格的判断）
					var checkMsg struct {
						Code int    `json:"code"`
						Msg  string `json:"msg"`
					}
					if err := json.Unmarshal(rawBody, &checkMsg); err == nil && checkMsg.Msg != "" {
						// 已经是统一格式，直接透传
						w.WriteHeader(status)
						_, _ = w.Write(rawBody)
						return
					}
				}
			}

			// 根据状态码判断成功/失败
			if status >= http.StatusBadRequest {
				// 错误响应：尝试解析错误信息
				var errData interface{}
				if err := json.Unmarshal(rawBody, &errData); err != nil {
					errData = string(rawBody)
				}

				// 尝试提取错误消息
				msg := http.StatusText(status)
				if m, ok := errData.(map[string]interface{}); ok {
					if emsg, ok := m["message"].(string); ok && emsg != "" {
						msg = emsg
					} else if emsg, ok := m["error"].(string); ok && emsg != "" {
						msg = emsg
					}
				}

				// 根据 HTTP 状态码映射到业务错误码
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

			// 成功响应：将原始数据包装到 data 字段
			var data interface{}
			if err := json.Unmarshal(rawBody, &data); err != nil {
				// 如果不是 JSON，作为字符串返回
				data = string(rawBody)
			}

			// 返回统一格式的成功响应
			response.OkWithData(w, data)
		}
	}
}
