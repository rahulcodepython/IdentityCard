package utils

import "net/http"

// APIError is the HTTP-shaped error type returned by services and controllers.
type APIError struct {
    Status  int               `json:"status"`
    Code    string            `json:"code,omitempty"`
    Message string            `json:"message"`
    Err     error             `json:"-"`
    Fields  map[string]string `json:"fields,omitempty"`
}

func (e *APIError) Error() string {
    return e.Message
}

func (e *APIError) Unwrap() error {
    return e.Err
}

func NewError(status int, message string, err error) *APIError {
    code := "error"
    switch status {
    case http.StatusBadRequest:
        code = "bad_request"
    case http.StatusUnauthorized:
        code = "unauthorized"
    case http.StatusForbidden:
        code = "forbidden"
    case http.StatusNotFound:
        code = "not_found"
    case http.StatusConflict:
        code = "conflict"
    case http.StatusUnprocessableEntity:
        code = "validation_error"
    case http.StatusTooManyRequests:
        code = "too_many_requests"
    case http.StatusInternalServerError:
        code = "internal_error"
    }
    return &APIError{
        Status:  status,
        Code:    code,
        Message: message,
        Err:     err,
    }
}

func ErrNotFound(message string, err error) *APIError {
    if message == "" {
        message = "Resource not found."
    }
    return NewError(http.StatusNotFound, message, err)
}

func ErrForbidden(message string, err error) *APIError {
    if message == "" {
        message = "Access denied."
    }
    return NewError(http.StatusForbidden, message, err)
}

func ErrBadRequest(message string, err error) *APIError {
    if message == "" {
        message = "Invalid request."
    }
    return NewError(http.StatusBadRequest, message, err)
}

func ErrConflict(message string, err error) *APIError {
    if message == "" {
        message = "Resource conflict or constraint violation."
    }
    return NewError(http.StatusConflict, message, err)
}

func ErrInternal(message string, err error) *APIError {
    if message == "" {
        message = "An unexpected error occurred."
    }
    return NewError(http.StatusInternalServerError, message, err)
}

func ErrValidation(fields map[string]string) *APIError {
    return &APIError{
        Status:  http.StatusUnprocessableEntity,
        Code:    "validation_error",
        Message: "One or more fields are invalid.",
        Fields:  fields,
    }
}

func ErrUnauthorized(message string) *APIError {
    if message == "" {
        message = "Authentication required."
    }
    return NewError(http.StatusUnauthorized, message, nil)
}

func ErrTooManyRequests(message string, err error) *APIError {
    if message == "" {
        message = "Too many requests. Please try again later."
    }
    return NewError(http.StatusTooManyRequests, message, err)
}
