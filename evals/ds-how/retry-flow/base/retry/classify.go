package retry

import "errors"

// Permanent marks an error that no retry can fix; Do returns it at once.
type Permanent struct{ Err error }

func (p Permanent) Error() string { return p.Err.Error() }
func (p Permanent) Unwrap() error { return p.Err }

func retryable(err error) bool {
	var perm Permanent
	return !errors.As(err, &perm)
}
