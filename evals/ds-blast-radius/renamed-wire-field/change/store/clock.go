package store

import "time"

// Now returns the current time; millisecond precision so ages sort stably.
var Now = func() int64 { return time.Now().UnixMilli() }
