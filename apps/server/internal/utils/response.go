package utils

import (
    "identitycard-server/internal/generic"

    "github.com/gofiber/fiber/v2"
)

// Response is an alias for the canonical wire envelope.
type Response = generic.Response[any]

// json is the canonical response envelope used by every handler (success,
// via OK/Created below) and by the central ErrorHandler (errors.go/
// error.handler.go) for every failure. It's the one place the wire shape is
// defined.
func json[T interface{}](c *fiber.Ctx, status int, success bool, message string, data T, err error) error {
    var errStr string
    if err != nil {
        errStr = err.Error()
        c.Locals("handler_error", err)
    } else if status >= 400 {
        errStr = message
    }
    if status >= 400 {
        c.Locals("handler_error_msg", message)
    }
    body := generic.Response[T]{
        Success: success,
        Message: message,
        Data:    data,
        Error:   errStr,
    }
    return c.Status(status).JSON(body)
}

// OK writes a success payload with HTTP 200 OK (or custom 2xx status) in the standard envelope.
// Supports both OK(c, message, data) and OK(c, status, data).
func OK[T interface{}](c *fiber.Ctx, msgOrStatus any, data ...T) error {
    status := fiber.StatusOK
    message := "success"

    switch v := msgOrStatus.(type) {
    case int:
        status = v
    case string:
        message = v
    }

    var payload T
    if len(data) > 0 {
        payload = data[0]
    }

    return json(c, status, true, message, payload, nil)
}

// OKEmpty writes a success payload with no data.
func OKEmpty(c *fiber.Ctx, message string) error {
    return json[*struct{}](c, fiber.StatusOK, true, message, nil, nil)
}

// Created writes a success payload with HTTP 201 Created.
// Supports both Created(c, message, data) and Created(c, data).
func Created[T interface{}](c *fiber.Ctx, msgOrData any, optionalData ...T) error {
    if msg, ok := msgOrData.(string); ok && len(optionalData) > 0 {
        return json(c, fiber.StatusCreated, true, msg, optionalData[0], nil)
    }
    var d any = msgOrData
    if len(optionalData) > 0 {
        d = optionalData[0]
    }
    return json(c, fiber.StatusCreated, true, "created", d, nil)
}

// OKWithMessage writes a success payload with a custom status and message.
func OKWithMessage(c *fiber.Ctx, status int, message string, data any) error {
    return json(c, status, true, message, data, nil)
}
