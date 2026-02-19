// Package response provides unified HTTP responses.
package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/addls/go-base/pkg/errcode"
)

// Response is the unified response structure.
type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"traceId,omitempty"`
}

// PageData represents paginated data.
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// Ok returns a success response (no data).
func Ok(w http.ResponseWriter) {
	httpx.OkJson(w, &Response{
		Code: errcode.OK.Code,
		Msg:  errcode.OK.Msg,
	})
}

// OkWithData returns a success response with data.
func OkWithData(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, &Response{
		Code: errcode.OK.Code,
		Msg:  errcode.OK.Msg,
		Data: data,
	})
}

// OkWithMsg returns a success response with a custom message.
func OkWithMsg(w http.ResponseWriter, msg string) {
	httpx.OkJson(w, &Response{
		Code: errcode.OK.Code,
		Msg:  msg,
	})
}

// OkWithPage returns a paginated success response.
func OkWithPage(w http.ResponseWriter, list interface{}, total int64, page, pageSize int) {
	httpx.OkJson(w, &Response{
		Code: errcode.OK.Code,
		Msg:  errcode.OK.Msg,
		Data: &PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Error returns an error response.
func Error(w http.ResponseWriter, err error) {
	e := errcode.FromError(err)
	httpx.WriteJson(w, e.GetHTTPCode(), &Response{
		Code: e.Code,
		Msg:  e.Msg,
	})
}

// ErrorWithMsg returns an error response with a custom message.
func ErrorWithMsg(w http.ResponseWriter, err *errcode.Error, msg string) {
	httpx.WriteJson(w, err.GetHTTPCode(), &Response{
		Code: err.Code,
		Msg:  msg,
	})
}

// ErrorWithCode returns an error response with the specified code and message.
func ErrorWithCode(w http.ResponseWriter, code int, msg string) {
	httpx.WriteJson(w, http.StatusOK, &Response{
		Code: code,
		Msg:  msg,
	})
}

// UnauthorizedCallback is for rest.WithUnauthorizedCallback (e.g. HTTP service JWT via rest.WithJwt).
// It responds with HTTP 401 and the unified response format { code, msg }.
func UnauthorizedCallback(w http.ResponseWriter, _ *http.Request, _ error) {
	Error(w, errcode.ErrUnauthorized)
}

// ErrorInvalidParam returns an invalid-parameter error response (used for parameter parsing failures).
// It converts a generic error into errcode.ErrInvalidParam.
func ErrorInvalidParam(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	// Use ErrInvalidParam while keeping the original error message.
	httpx.WriteJson(w, errcode.ErrInvalidParam.GetHTTPCode(), &Response{
		Code: errcode.ErrInvalidParam.Code,
		Msg:  err.Error(), // Keep original message for debugging.
	})
}

// ----- TraceID variants -----

// OkWithTrace returns a success response with TraceID.
func OkWithTrace(w http.ResponseWriter, data interface{}, traceID string) {
	httpx.OkJson(w, &Response{
		Code:    errcode.OK.Code,
		Msg:     errcode.OK.Msg,
		Data:    data,
		TraceID: traceID,
	})
}

// ErrorWithTrace returns an error response with TraceID.
func ErrorWithTrace(w http.ResponseWriter, err error, traceID string) {
	e := errcode.FromError(err)
	httpx.WriteJson(w, e.GetHTTPCode(), &Response{
		Code:    e.Code,
		Msg:     e.Msg,
		TraceID: traceID,
	})
}

// ----- go-zero handler helpers -----

// HandleResult handles handler results in a unified way.
// Usage: response.HandleResult(w, resp, err)
func HandleResult(w http.ResponseWriter, resp interface{}, err error) {
	if err != nil {
		Error(w, err)
	} else {
		OkWithData(w, resp)
	}
}

// HandleResultWithPage handles paginated results in a unified way.
func HandleResultWithPage(w http.ResponseWriter, list interface{}, total int64, page, pageSize int, err error) {
	if err != nil {
		Error(w, err)
	} else {
		OkWithPage(w, list, total, page, pageSize)
	}
}
