package colony

import (
	"context"
	"time"
)

func waitForAction(ctx context.Context, duration time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
