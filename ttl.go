// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"fmt"
	"math"
	"math/rand/v2"
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

// TTLJitterer produces DNS TTL values that are jittered by a certain percent. This
// type is immutable and safe for concurrent use.
//
// The zero value for this type simple returns zero (0) for all TTLs. Use NewTTLJitterer
// to create a more interesting TTLJitterer.
type TTLJitterer struct {
	// base is the lowest TTL this jitterer will choose.
	base uint32

	// choose is the range of values this jitterer will select.
	// A random value in this range is selected and added to base.
	choose uint32
}

// NewTTLJitterer creates a jitterer for DNS TTLs.
//
// Value is the original DNS TTL value. If this value is zero, it is always
// returned as is.
//
// Percent is the percentage above and below the Value. A random value will be
// selected in this range. For example, with a base of 3600 and a percent of 0.1,
// the TTLs will range from 3240 (10% below) to 3960 (10% above). If the Percent is
// 0.0, the value is always returned as is. If the percent is negative or greater
// than or equal to 1.0, an error is returned.
func NewTTLJitterer(value uint32, percent float32) (j *TTLJitterer, err error) {
	j = new(TTLJitterer)

	switch {
	case percent < 0.0 || percent >= 1.0:
		err = fmt.Errorf("invalid jitter percent %g", percent)

	case percent == 0.0 || value == 0:
		j.base = value

	default:
		j.base = uint32((1.0 - percent) * float32(value))
		max := uint32((1.0 + percent) * float32(value))

		// rand.Uint32N chooses values in the half open range [0, n), and we want the full range
		j.choose = (max - j.base) + 1
	}

	return
}

// Range returns the inclusive range of TTLs this jitterer will generate.
func (j *TTLJitterer) Range() (lo, hi uint32) {
	if j.choose == 0 {
		return j.base, j.base
	}

	return j.base, j.base + j.choose - 1
}

// TTL chooses a random value in the range [(1-percent)*value, (1+percent)*value]. If value
// or percent passed to NewTTLJitterer was zero, this method returns the TTL value as is.
func (j *TTLJitterer) TTL() uint32 {
	if j.choose == 0 {
		return j.base
	}

	return j.base + rand.Uint32N(j.choose)
}
