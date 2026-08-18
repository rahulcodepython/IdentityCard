package httpx

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

type envelope struct {
	Data any       `json:"data,omitempty"`
	Err  *APIError `json:"error,omitempty"`
}

// OK writes a typed success payload in the standard {"data": ...} envelope.
func OK(c *fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(envelope{Data: data})
}

// ErrorHandler is installed as the Fiber app's central error handler so
// every handler can just `return httpx.ErrNotFound("event")` etc. and get
// a consistently shaped {"error": {...}} response with the right status.
func ErrorHandler(c *fiber.Ctx, err error) error {
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return c.Status(apiErr.Status).JSON(envelope{Err: apiErr})
	}

	if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
		return c.Status(fiberErr.Code).JSON(envelope{
			Err: &APIError{Code: "request_error", Message: fiberErr.Message},
		})
	}

	log.Printf("unhandled error: %v", err)
	internal := ErrInternal()
	return c.Status(internal.Status).JSON(envelope{Err: internal})
}
