package middlewares

import (
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/gofiber/fiber/v2"
)

// LoggerMiddleware provides structured HTTP request logging using log/slog.
func LoggerMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        err := c.Next()

        latency := time.Since(start)
        status := c.Response().StatusCode()

        if err != nil {
            c.Locals("handler_error", err)
        }

        handlerErr := c.Locals("handler_error")
        routePath := "-"
        if r := c.Route(); r != nil {
            routePath = r.Path
        }

        if status >= 400 || handlerErr != nil {
            var errDetail string
            if handlerErr != nil {
                if e, ok := handlerErr.(error); ok {
                    errDetail = e.Error()
                } else {
                    errDetail = fmt.Sprintf("%v", handlerErr)
                }
            }
            if errDetail == "" {
                errDetail = fmt.Sprintf("HTTP %d %s", status, http.StatusText(status))
            }

            claims := Claims(c)
            var orgIDStr string
            if claims != nil {
                orgIDStr = claims.OrganizationID.String()
            }

            slog.Error("api failure",
                "method", c.Method(),
                "path", c.Path(),
                "route", routePath,
                "status", status,
                "latency_ms", latency.Milliseconds(),
                "ip", c.IP(),
                "org_id", orgIDStr,
                "error", errDetail,
            )
        } else {
            slog.Info("api request",
                "method", c.Method(),
                "path", c.Path(),
                "status", status,
                "latency_ms", latency.Milliseconds(),
                "ip", c.IP(),
            )
        }

        return err
    }
}
