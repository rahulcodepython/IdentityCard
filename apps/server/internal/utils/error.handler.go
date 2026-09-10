package utils

import (
    "context"
    "errors"

    "github.com/gofiber/fiber/v2"
)

// ErrorHandler is the single central place every failure response is
// rendered from — handlers/services never write a status code or JSON body
// themselves; they return an *APIError (see errors.go) and this renders it
// through the standard response envelope (json, in response.go). Wired into
// fiber.Config{ErrorHandler: ...} at construction (cmd/server/main.go).
func ErrorHandler(c *fiber.Ctx, err error) error {
    var apiErr *APIError
    if errors.As(err, &apiErr) {
        if errors.Is(apiErr.Err, context.Canceled) {
            c.Locals("handler_error", apiErr.Err)
            return json[*struct{}](c, 499, false, "Request was canceled.", nil, apiErr.Err)
        }
        if errors.Is(apiErr.Err, context.DeadlineExceeded) {
            c.Locals("handler_error", apiErr.Err)
            return json[*struct{}](c, fiber.StatusGatewayTimeout, false, "Request timed out.", nil, apiErr.Err)
        }
        payload := apiErr.Data
        if payload == nil && len(apiErr.Fields) > 0 {
            payload = apiErr.Fields
        }
        return json(c, apiErr.Status, false, apiErr.Message, payload, apiErr.Err)
    }

    if errors.Is(err, context.Canceled) {
        c.Locals("handler_error", err)
        return json[*struct{}](c, 499, false, "Request was canceled.", nil, err)
    }
    if errors.Is(err, context.DeadlineExceeded) {
        c.Locals("handler_error", err)
        return json[*struct{}](c, fiber.StatusGatewayTimeout, false, "Request timed out.", nil, err)
    }

    code := fiber.StatusInternalServerError
    if fe, ok := err.(*fiber.Error); ok {
        code = fe.Code
    }

    c.Locals("handler_error", err)

    if code == fiber.StatusNotFound {
        return json[*struct{}](c, fiber.StatusNotFound, false, "Requested resource not found.", nil, err)
    }

    // Anything else (a framework-level error such as a body-size limit, or a
    // recovered panic) falls back to a generic message — the raw error text
    // is logged via handler_error above but never leaked to the client.
    return json[*struct{}](c, code, false, "An unexpected error occurred.", nil, nil)
}
