// Package jobs runs periodic background sweeps — billing (billing.go) and
// recurring-event horizon extension (recurrence.go) — that don't belong
// on any single HTTP request. A prototype-simple ticker goroutine, not a
// dedicated cron dependency — there's no case yet for more infrastructure
// than that.
package jobs

import (
	"context"
	"log"
	"time"
)

// Task is one independent unit of periodic work.
type Task func(context.Context) error

// Run executes every task once immediately, then again every interval,
// until ctx is canceled. Each task's panics/errors are recovered and
// logged independently — mirrors the fire-and-forget goroutine in
// events.Handler.Publish (its card-emailing call), which has the same
// requirement: this goroutine outlives any single request and isn't
// covered by Fiber's recover middleware, so an unrecovered panic here
// would crash the whole process. Intended to be started as
// `go jobs.Run(ctx, 24*time.Hour, ...)` from main.go.
func Run(ctx context.Context, interval time.Duration, tasks ...Task) {
	runAll(ctx, tasks)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runAll(ctx, tasks)
		}
	}
}

func runAll(ctx context.Context, tasks []Task) {
	for i, task := range tasks {
		runOne(ctx, i, task)
	}
}

func runOne(ctx context.Context, index int, task Task) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("jobs: panic in task %d: %v", index, r)
		}
	}()
	if err := task(ctx); err != nil {
		log.Printf("jobs: task %d failed: %v", index, err)
	}
}
