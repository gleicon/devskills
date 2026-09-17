package retry

import (
	"context"
	"fmt"
	"time"
)

// Do calls fn until it returns nil, a Permanent error, or the policy is exhausted.
func Do(ctx context.Context, p Policy, fn func() error) error {
	var last error
	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		last = fn()
		if last == nil || !retryable(last) {
			return last
		}
		if attempt == p.MaxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(p.Delay(attempt)):
		}
	}
	return fmt.Errorf("retry: %d attempts: %w", p.MaxAttempts, last)
}
