// Package response 提供统一的 HTTP 响应
package response

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/addls/go-base/pkg/errcode"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"traceId,omitempty"`
}

// PageData 分页数据
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// Ok 成功响应（无数据）
func Ok(w http.ResponseWriter) {
	httpx.OkJson(w, &Response{
		Code: errcode.OK.Code,
		Msg:  errcode.OK.Msg,
	})
}

// OkWithData 成功响应（带数据）
func OkWithData(w http.ResponseWriter, data interface{}) {
	httpx.OkJson(w, &Response{
		Code: errcode.OK.Code,
		Msg:  errcode.OK.Msg,
		Data: data,
	})
}

// OkWithMsg 成功响应（自定义消息）
func OkWithMsg(w http.ResponseWriter, msg string) {
	httpx.OkJson(w, &Response{
		Code: errcode.OK.Code,
		Msg:  msg,
	})
}

// OkWithPage 分页成功响应
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

// Error 错误响应
func Error(w http.ResponseWriter, err error) {
	e := errcode.FromError(err)
	httpx.WriteJson(w, e.GetHTTPCode(), &Response{
		Code: e.Code,
		Msg:  e.Msg,
	})
}

// ErrorWithMsg 错误响应（自定义消息）
func ErrorWithMsg(w http.ResponseWriter, err *errcode.Error, msg string) {
	httpx.WriteJson(w, err.GetHTTPCode(), &Response{
		Code: err.Code,
		Msg:  msg,
	})
}

// ErrorWithCode 错误响应（指定错误码和消息）
func ErrorWithCode(w http.ResponseWriter, code int, msg string) {
	httpx.WriteJson(w, http.StatusOK, &Response{
		Code: code,
		Msg:  msg,
	})
}

// ----- 带 TraceID 的版本 -----

// OkWithTrace 成功响应（带 TraceID）
func OkWithTrace(w http.ResponseWriter, data interface{}, traceID string) {
	httpx.OkJson(w, &Response{
		Code:    errcode.OK.Code,
		Msg:     errcode.OK.Msg,
		Data:    data,
		TraceID: traceID,
	})
}

// ErrorWithTrace 错误响应（带 TraceID）
func ErrorWithTrace(w http.ResponseWriter, err error, traceID string) {
	e := errcode.FromError(err)
	httpx.WriteJson(w, e.GetHTTPCode(), &Response{
		Code:    e.Code,
		Msg:     e.Msg,
		TraceID: traceID,
	})
}

// ----- go-zero handler 封装 -----

// HandleResult 统一处理 handler 结果
// 用法: response.HandleResult(w, resp, err)
func HandleResult(w http.ResponseWriter, resp interface{}, err error) {
	if err != nil {
		Error(w, err)
	} else {
		OkWithData(w, resp)
	}
}

// HandleResultWithPage 统一处理分页结果
func HandleResultWithPage(w http.ResponseWriter, list interface{}, total int64, page, pageSize int, err error) {
	if err != nil {
		Error(w, err)
	} else {
		OkWithPage(w, list, total, page, pageSize)
	}
}
