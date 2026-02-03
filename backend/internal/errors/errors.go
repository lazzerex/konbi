package errors

import (
	"fmt"
	"net/http"
)

// app error represents a domain-specific error
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
}

// implement error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// predefined error constructors
func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Code:       "BAD_REQUEST",
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    message,
		StatusCode: http.StatusNotFound,
		Err:        nil,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Err:        nil,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    message,
		StatusCode: http.StatusForbidden,
		Err:        nil,
	}
}

func NewRateLimitError() *AppError {
	return &AppError{
		Code:       "RATE_LIMIT_EXCEEDED",
		Message:    "rate limit exceeded",
		StatusCode: http.StatusTooManyRequests,
		Err:        nil,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    message,
		StatusCode: http.StatusConflict,
		Err:        nil,
	}
}

// validation errors
func NewFileTooLargeError(maxSize int64) *AppError {
	return &AppError{
		Code:       "FILE_TOO_LARGE",
		Message:    fmt.Sprintf("file size exceeds %dMB limit", maxSize/1024/1024),
		StatusCode: http.StatusBadRequest,
		Err:        nil,
	}
}

func NewFileTypeNotAllowedError() *AppError {
	return &AppError{
		Code:       "FILE_TYPE_NOT_ALLOWED",
		Message:    "file type not allowed",
		StatusCode: http.StatusBadRequest,
		Err:        nil,
	}
}

func NewContentTooLargeError() *AppError {
	return &AppError{
		Code:       "CONTENT_TOO_LARGE",
		Message:    "content too large",
		StatusCode: http.StatusBadRequest,
		Err:        nil,
	}
}
