package utils

import "net/http"

// APIError is the only error type handlers should return up the stack.
// The central Fiber error handler (see ErrorHandler) knows how to render
// it; anything else is treated as an unexpected 500.
type APIError struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *APIError) Error() string { return e.Message }

func NewError(status int, code, message string) *APIError {
	return &APIError{Status: status, Code: code, Message: message}
}

func ErrValidation(fields map[string]string) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    "validation_error",
		Message: "one or more fields are invalid",
		Fields:  fields,
	}
}

func ErrUnauthorized(message string) *APIError {
	if message == "" {
		message = "authentication required"
	}
	return NewError(http.StatusUnauthorized, "unauthorized", message)
}

func ErrForbidden(message string) *APIError {
	if message == "" {
		message = "you do not have permission to perform this action"
	}
	return NewError(http.StatusForbidden, "forbidden", message)
}

func ErrNotFound(resource string) *APIError {
	return NewError(http.StatusNotFound, "not_found", resource+" not found")
}

func ErrConflict(message string) *APIError {
	return NewError(http.StatusConflict, "conflict", message)
}

func ErrInternal() *APIError {
	return NewError(http.StatusInternalServerError, "internal_error", "something went wrong")
}
