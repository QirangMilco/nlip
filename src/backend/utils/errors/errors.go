package errors

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
)

// ErrorCode 错误码类型
type ErrorCode int

const (
    // 通用错误码
    ErrBadRequest ErrorCode = iota + 40000
    ErrUnauthorized
    ErrForbidden
    ErrNotFound
    ErrInternalServer

    // 认证相关错误码
    ErrInvalidCredentials
    ErrTokenExpired
    ErrTokenInvalid

    // 业务相关错误码
    ErrSpaceNotFound
    ErrClipNotFound
    ErrNoPermission
    ErrFileTooLarge
    ErrInvalidFileType
)

// NlipError 应用错误类型
type NlipError struct {
    Code    ErrorCode `json:"code"`
    Message string    `json:"message"`
    Details string    `json:"details,omitempty"`
}

func (e *NlipError) Error() string {
    if e.Details != "" {
        return fmt.Sprintf("%s: %s", e.Message, e.Details)
    }
    return e.Message
}

// NewNlipError 创建新的应用错误
func NewNlipError(code ErrorCode, message string) *NlipError {
    return &NlipError{
        Code:    code,
        Message: message,
    }
}

// WithDetails 添加错误详情
func (e *NlipError) WithDetails(details string) *NlipError {
    e.Details = details
    return e
}

// ToFiberError 转换为Fiber错误
func (e *NlipError) ToFiberError() *fiber.Error {
    return fiber.NewError(int(e.Code), e.Error())
}

// FromFiberError 从Fiber错误创建应用错误
func FromFiberError(err *fiber.Error) *NlipError {
    return &NlipError{
        Code:    ErrorCode(err.Code),
        Message: err.Message,
    }
} 