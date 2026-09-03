// Package service 实现业务规则与 DTO<->模型转换，handler 只做 HTTP 编解码。
package service

import (
	"errors"
	"net/http"
)

// Error 业务错误，携带 HTTP 状态码与错误码。
type Error struct {
	Status  int
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

// ErrNotFound 资源不存在。
var ErrNotFound = &Error{Status: http.StatusNotFound, Code: "NOT_FOUND", Message: "资源不存在"}

// BadRequest 返回 400 错误。
func BadRequest(msg string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: msg}
}

// Unprocessable 返回 422 业务校验错误。
func Unprocessable(msg string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: "VALIDATION_ERROR", Message: msg}
}

// Conflict 返回 409 冲突错误（如被引用无法删除）。
func Conflict(msg string) *Error {
	return &Error{Status: http.StatusConflict, Code: "CONFLICT", Message: msg}
}

// Internal 包装内部错误为 500。
func Internal(err error) *Error {
	return &Error{Status: http.StatusInternalServerError, Code: "INTERNAL_ERROR", Message: err.Error()}
}

// AsError 将任意 error 归一化为 *Error。
func AsError(err error) *Error {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}
