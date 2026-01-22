// Package errcode 提供统一的错误码定义
package errcode

import (
	"fmt"
	"net/http"
)

// Error 统一错误结构
type Error struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	HTTPCode int    `json:"-"`
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

// WithMsg 返回带新消息的错误（不修改原错误）
func (e *Error) WithMsg(msg string) *Error {
	return &Error{
		Code:     e.Code,
		Msg:      msg,
		HTTPCode: e.HTTPCode,
	}
}

// GetHTTPCode 获取 HTTP 状态码
func (e *Error) GetHTTPCode() int {
	if e.HTTPCode > 0 {
		return e.HTTPCode
	}
	return http.StatusOK
}

// New 创建错误码
func New(code int, msg string) *Error {
	return &Error{
		Code:     code,
		Msg:      msg,
		HTTPCode: http.StatusOK,
	}
}

// NewWithHTTP 创建带 HTTP 状态码的错误
func NewWithHTTP(code int, msg string, httpCode int) *Error {
	return &Error{
		Code:     code,
		Msg:      msg,
		HTTPCode: httpCode,
	}
}

// IsError 判断是否为指定错误码
func IsError(err error, target *Error) bool {
	if err == nil || target == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == target.Code
	}
	return false
}

// FromError 从 error 转为 *Error
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return ErrInternal.WithMsg(err.Error())
}

// Code 获取错误码
func Code(err error) int {
	if err == nil {
		return OK.Code
	}
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return ErrInternal.Code
}

// Msg 获取错误信息
func Msg(err error) string {
	if err == nil {
		return OK.Msg
	}
	if e, ok := err.(*Error); ok {
		return e.Msg
	}
	return err.Error()
}
