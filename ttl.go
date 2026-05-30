package hashy

import (
	"math"
	"time"
)

// DurationToSeconds converts a time.Duration into a uint32 number of seconds
// as required by the TTL field of DNS records. Fractional seconds are rounded to the nearest
// second. Durations that are too large are truncated to a uint32.
func DurationToSeconds(v time.Duration) uint32 {
	if v <= 0 {
		return 0
	}

	seconds := uint32(math.Round(v.Seconds()))
	return max(seconds, 1)
}

// SecondsToDuration converts a uint32 TTL in seconds, as found in a DNS record, into a time.Duration.
func SecondsToDuration(v uint32) time.Duration {
	return time.Duration(v) * time.Second
}
