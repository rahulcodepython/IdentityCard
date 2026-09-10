package retry

import (
    "log/slog"
    "time"
)

// Connect calls fn up to maxAttempts times with a fixed interval between attempts,
// logging each failed attempt under the given label. It returns nil on the first
// success, or the last error if all attempts are exhausted.
func Connect(label string, maxAttempts int, interval time.Duration, fn func() error) error {
    var lastErr error
    for attempt := 1; attempt <= maxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        lastErr = err
        if attempt < maxAttempts {
            slog.Warn("connection attempt failed, retrying",
                "component", label,
                "attempt", attempt,
                "max_attempts", maxAttempts,
                "retry_in", interval.String(),
                "error", err,
            )
            time.Sleep(interval)
        }
    }
    return lastErr
}
