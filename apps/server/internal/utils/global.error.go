package utils

import (
    "context"
    "errors"

    "github.com/gofiber/fiber/v2"
)

// ErrorHandler is Fiber's global error handler. It intercepts all returned
// errors and sends the canonical error response using the Respond function.
func ErrorHandler(c *fiber.Ctx, err error) error {
    if errors.Is(err, context.Canceled) {
        return Respond(c, 499, false, "Request was canceled.", any(nil), err)
    }
    if errors.Is(err, context.DeadlineExceeded) {
        return Respond(c, fiber.StatusGatewayTimeout, false, "Request timed out.", any(nil), err)
    }

    code := fiber.StatusInternalServerError
    message := "An unexpected error occurred."

    var fe *fiber.Error
    if errors.As(err, &fe) {
        code = fe.Code
        message = fe.Message
    } else if errors.Is(err, fiber.ErrNotFound) {
        code = fiber.StatusNotFound
        message = "Requested resource not found."
    }

    return Respond(c, code, false, message, any(nil), err)
}
