package utils

import "github.com/gofiber/fiber/v2"

// APIError is the only error type any handler/service returns up the call
// stack. Nothing below the HTTP layer writes a status code or JSON body
// directly — it constructs one of these and returns it, and the central
// ErrorHandler (error.handler.go) renders it through the standard response
// envelope. Err carries the underlying cause for logging only; it is never
// serialized to the client. Data is optional and nil for almost every
// caller — it exists for the rare non-2xx response that still needs to
// carry a payload (e.g. a 503 health check reporting which dependencies
// are down), set directly on the struct literal rather than via a
// dedicated constructor.
type APIError struct {
    Status  int               `json:"status"`
    Code    string            `json:"code,omitempty"`
    Message string            `json:"message"`
    Data    interface{}       `json:"data,omitempty"`
    Err     error             `json:"-"`
    Fields  map[string]string `json:"fields,omitempty"`
}

func (e *APIError) Error() string { return e.Message }
func (e *APIError) Unwrap() error { return e.Err }

func NewError(status int, message string, err error) *APIError {
    return &APIError{Status: status, Message: message, Err: err}
}

func ErrBadRequest(message string, err error) *APIError {
    return NewError(fiber.StatusBadRequest, message, err)
}

// ErrValidation is for request-shape/struct-tag validation failures —
// mirrors the project's existing use of 422 (not 400) for that case.
func ErrValidation(target any, err ...error) *APIError {
    switch v := target.(type) {
    case string:
        var cause error
        if len(err) > 0 {
            cause = err[0]
        }
        return &APIError{
            Status:  fiber.StatusUnprocessableEntity,
            Message: v,
            Err:     cause,
        }
    case map[string]string:
        var cause error
        if len(err) > 0 {
            cause = err[0]
        }
        return &APIError{
            Status:  fiber.StatusUnprocessableEntity,
            Message: "Validation failed.",
            Data:    v,
            Fields:  v,
            Err:     cause,
        }
    default:
        var cause error
        if len(err) > 0 {
            cause = err[0]
        }
        return &APIError{
            Status:  fiber.StatusUnprocessableEntity,
            Message: "Validation failed.",
            Err:     cause,
        }
    }
}

func ErrUnauthorized(message string, err ...error) *APIError {
    var cause error
    if len(err) > 0 {
        cause = err[0]
    }
    return NewError(fiber.StatusUnauthorized, message, cause)
}

func ErrForbidden(message string, err error) *APIError {
    return NewError(fiber.StatusForbidden, message, err)
}

func ErrNotFound(message string, err error) *APIError {
    return NewError(fiber.StatusNotFound, message, err)
}

func ErrConflict(message string, err error) *APIError {
    return NewError(fiber.StatusConflict, message, err)
}

func ErrTooManyRequests(message string, err error) *APIError {
    return NewError(fiber.StatusTooManyRequests, message, err)
}

func ErrInternal(message string, err error) *APIError {
    return NewError(fiber.StatusInternalServerError, message, err)
}
