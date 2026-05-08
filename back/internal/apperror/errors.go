package apperror

import "net/http"

type AppError struct {
	HTTPStatus int         `json:"-"`
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(status int, code, message string, details interface{}) *AppError {
	return &AppError{HTTPStatus: status, Code: code, Message: message, Details: details}
}

func BadRequest(code, message string, details interface{}) *AppError {
	return New(http.StatusBadRequest, code, message, details)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, "unauthorized", message, nil)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, "forbidden", message, nil)
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, "not_found", message, nil)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, "conflict", message, nil)
}

func Internal(message string) *AppError {
	return New(http.StatusInternalServerError, "internal_error", message, nil)
}
