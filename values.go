// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"iter"
	"slices"
)

// Values is a slice of sortable, dedupable values.
type Values[V comparable] []V

// Add appends more values. After adding several values,
// use Normalize to dedupe and sort this collection again.
func (vs *Values[V]) Add(more ...V) {
	*vs = slices.Grow(*vs, len(more))
	*vs = append(*vs, more...)
}

// Append is like Add, but is appends to this Deduper and returns
// the appended Deduper rather than adding in-place.
func (vs Values[V]) Append(more ...V) Values[V] {
	vs = slices.Grow(vs, len(more))
	vs = append(vs, more...)
	return vs
}

// Dedupe removes any duplicates from this set. After this method is
// called, the order of elements may have shifted.
func (vs *Values[V]) Dedupe() {
	if len(*vs) < 2 {
		return
	}

	var deduper Deduper[V]
	deduper.Add((*vs)...)

	if deduper.Len() < vs.Len() {
		// zero out the elements we no longer need, but retain the
		// slice storage for future use.
		*vs = slices.Delete(*vs, deduper.Len(), vs.Len())
		*vs = deduper.AppendTo((*vs)[:])
	}
}

// SortFunc in-place sorts this set using the supplied comparison function.
func (vs Values[V]) SortFunc(cmp func(V, V) int) {
	slices.SortFunc(vs, cmp)
}

// Len returns the count of elements in this set.
func (vs Values[V]) Len() int {
	return len(vs)
}

// All returns a sequence of this set's values in the order they appear.
func (vs Values[V]) All() iter.Seq[V] {
	return slices.Values(vs)
}
