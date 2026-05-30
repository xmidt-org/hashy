// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"maps"
	"slices"
)

// Deduper is a utility type used to dedupe sequences of comparable values.
type Deduper[V comparable] map[V]struct{}

// Len returns the count of unique values.
func (d Deduper[V]) Len() int {
	return len(d)
}

// Clear wipes out this Deduper.
func (d Deduper[V]) Clear() {
	clear(d)
}

// Add adds more values to dedupe. This method will create the map as needed.
func (d *Deduper[V]) Add(more ...V) {
	if *d == nil {
		*d = make(Deduper[V], len(more))
	}

	for _, v := range more {
		(*d)[v] = struct{}{}
	}
}

// Slice returns a slice of this Deduper's values. The returned
// slice will not be sorted, but it will be guaranteed not to have
// duplicate values.
func (d Deduper[V]) Slice() (deduped []V) {
	if d.Len() > 0 {
		deduped = slices.AppendSeq(
			make([]V, 0, d.Len()),
			maps.Keys(d),
		)
	}

	return
}

// AppendTo appends this Deduper's values to a given slice. The values
// are not appended in sorted order, but this method does guarantee that
// a given value will not be appended more than once.
func (d Deduper[V]) AppendTo(dst []V) []V {
	if d.Len() > 0 {
		dst = slices.AppendSeq(
			slices.Grow(dst, d.Len()),
			maps.Keys(d),
		)
	}

	return dst
}

// SortFunc is a convenience for taking the result of Slice and sorting
// it using the supplied comparison function.
func (d Deduper[V]) SortFunc(cmp func(V, V) int) (sorted []V) {
	sorted = d.Slice()
	slices.SortFunc(sorted, cmp)
	return
}
