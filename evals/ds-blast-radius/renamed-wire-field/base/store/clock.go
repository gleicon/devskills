package store

import "time"

// Now returns the current time as Unix seconds, the unit records are stored in.
var Now = func() int64 { return time.Now().Unix() }
