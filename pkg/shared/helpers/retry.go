package helpers

import (
	"context"
	"time"
)

func Retry(
	ctx context.Context,
	maxAttemts int,
	fn func() error,
) error {
	var err error

	delay := 2 * time.Second

	for attempt := 0; attempt < maxAttemts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if attempt == maxAttemts-1 {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay *= 2
		}
	}
	return err
}
