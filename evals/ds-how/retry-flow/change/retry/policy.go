package retry

import (
	"math/rand/v2"
	"time"
)

// Policy bounds a retry loop: how many attempts and how long to wait between them.
type Policy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// Delay returns the wait before the given attempt (1-based): exponential backoff
// with full jitter, so concurrent callers do not retry in lockstep.
func (p Policy) Delay(attempt int) time.Duration {
	d := min(p.BaseDelay<<(attempt-1), p.MaxDelay)
	return time.Duration(rand.Int64N(int64(d) + 1))
}
