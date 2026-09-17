package retry

import "time"

// Policy bounds a retry loop: how many attempts and how long to wait between them.
type Policy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// Delay returns the wait before the given attempt (1-based) using exponential backoff.
func (p Policy) Delay(attempt int) time.Duration {
	d := p.BaseDelay << (attempt - 1)
	if d > p.MaxDelay {
		return p.MaxDelay
	}
	return d
}
