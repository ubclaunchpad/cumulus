package common

import "time"

// UnixNow returns the current Unix time as a uint32. This represents the number
// of seconds elapsed since January 1, 1970 UTC.
func UnixNow() uint32 {
	return uint32(time.Now().Unix())
}
