package service

import (
	"cmp"
	"iter"
	"maps"
	"slices"
)

// collector is a mapping between a key and a set of values. The set of values
// will be deduped as (k, v) pairs are added.
type collector[K cmp.Ordered, V comparable] map[K]map[V]struct{}

// len returns the number of keys in this collector.
func (c collector[K, V]) len() int {
	return len(c)
}

// valuesLen returns the count of values for the given key.
func (c collector[K, V]) valuesLen(k K) int {
	return len(c[k])
}

// clear wipes out all keys and values from this collector.
func (c collector[K, V]) clear() {
	clear(c)
}

func (c *collector[K, V]) add(k K, v V) {
	if *c == nil {
		*c = collector[K, V]{
			k: {v: {}},
		}
	} else if members := (*c)[k]; members != nil {
		members[v] = struct{}{}
	} else {
		(*c)[k] = map[V]struct{}{
			v: {},
		}
	}
}

func (c *collector[K, V]) addSlice(k K, vslice []V) {
	var members map[V]struct{}
	if *c == nil {
		members = make(map[V]struct{}, len(vslice))
		*c = collector[K, V]{
			k: members,
		}
	} else if members = (*c)[k]; members == nil {
		members = make(map[V]struct{}, len(vslice))
		(*c)[k] = members
	}

	for _, v := range vslice {
		members[v] = struct{}{}
	}
}

// sortedKeys returns a sorted sequence of this collector's keys.
func (c collector[K, V]) sortedKeys() iter.Seq[K] {
	keys := slices.AppendSeq(
		make([]K, 0, c.len()),
		maps.Keys(c),
	)

	slices.SortFunc(keys, cmp.Compare)
	return slices.Values(keys)
}

// values returns an unsorted, deduped sequence of the values for the given key.
// If no such key exists, the returned sequence is empty.
func (c collector[K, V]) values(k K) iter.Seq[V] {
	return maps.Keys(c[k])
}

// sortedValuesSlice returns a slice of values for the given key. The values will be
// deduped and sorted using the given function.
//
// If no values exist for the given key, this method returns a nil slice.
func (c collector[K, V]) sortedValuesSlice(k K, cmp func(V, V) int) (values []V) {
	if members := c[k]; len(members) > 0 {
		values = slices.AppendSeq(
			make([]V, 0, len(members)),
			maps.Keys(members),
		)

		slices.SortFunc(values, cmp)
	}

	return
}
