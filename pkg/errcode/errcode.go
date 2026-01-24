// Package errcode provides unified error code definitions.
package errcode

import (
	"fmt"
	"net/http"
)

// Error is the unified error structure.
type Error struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	HTTPCode int    `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

// WithMsg returns a copy with a new message (does not modify the original error).
func (e *Error) WithMsg(msg string) *Error {
	return &Error{
		Code:     e.Code,
		Msg:      msg,
		HTTPCode: e.HTTPCode,
	}
}

// GetHTTPCode returns the HTTP status code.
func (e *Error) GetHTTPCode() int {
	if e.HTTPCode > 0 {
		return e.HTTPCode
	}
	return http.StatusOK
}

// New creates an error code.
func New(code int, msg string) *Error {
	return &Error{
		Code:     code,
		Msg:      msg,
		HTTPCode: http.StatusOK,
	}
}

// NewWithHTTP creates an error with an HTTP status code.
func NewWithHTTP(code int, msg string, httpCode int) *Error {
	return &Error{
		Code:     code,
		Msg:      msg,
		HTTPCode: httpCode,
	}
}

// IsError reports whether err matches the target error code.
func IsError(err error, target *Error) bool {
	if err == nil || target == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == target.Code
	}
	return false
}

// FromError converts an error into *Error.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return ErrInternal.WithMsg(err.Error())
}

// Code returns the error code.
func Code(err error) int {
	if err == nil {
		return OK.Code
	}
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return ErrInternal.Code
}

// Msg returns the error message.
func Msg(err error) string {
	if err == nil {
		return OK.Msg
	}
	if e, ok := err.(*Error); ok {
		return e.Msg
	}
	return err.Error()
}
