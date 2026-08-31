package utils

import (
    "errors"
    "log"

    "github.com/gofiber/fiber/v2"
)

type Response struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
    Error   any    `json:"error,omitempty"`
}

// OK writes a typed success payload in the standard {success, message, data} envelope.
func OK(c *fiber.Ctx, status int, data any) error {
    return c.Status(status).JSON(Response{
        Success: true,
        Message: "success",
        Data:    data,
    })
}

// OKWithMessage writes a typed success payload with a custom message.
func OKWithMessage(c *fiber.Ctx, status int, message string, data any) error {
    return c.Status(status).JSON(Response{
        Success: true,
        Message: message,
        Data:    data,
    })
}

// ErrorHandler is installed as Fiber's central error handler.
// It is the ONLY place a status code or JSON error body gets written.
func ErrorHandler(c *fiber.Ctx, err error) error {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        var errDetail any = apiErr.Message
        if apiErr.Err != nil {
            errDetail = apiErr.Err.Error()
        } else if len(apiErr.Fields) > 0 {
            errDetail = apiErr.Fields
        }
        return c.Status(apiErr.Status).JSON(Response{
            Success: false,
            Message: apiErr.Message,
            Data:    nil,
            Error:   errDetail,
        })
    }

    code := fiber.StatusInternalServerError
    var fiberErr *fiber.Error
    if errors.As(err, &fiberErr) {
        code = fiberErr.Code
    }

    log.Printf("unhandled error: %v", err)
    c.Locals("handler_error", err)

    return c.Status(code).JSON(Response{
        Success: false,
        Message: "An unexpected error occurred.",
        Data:    nil,
        Error:   err.Error(),
    })
}
