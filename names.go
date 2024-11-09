package hashy

import (
	"strings"
)

// normalizeName normalizes the given name.  Right now, all this
// function does is strings.ToLower.
func normalizeName(v string) string {
	return strings.ToLower(v)
}

// Names is an immutable set of service names.  Names are deduped
// and normalized. The zero value of this type is an empty set.
// Use NewNames to create a non-empty Names. A nil Names is valid,
// and is treated as empty by its methods.
type Names struct {
	n map[string]bool
}

// NewNames constructs an immutable Names set.
func NewNames(values ...string) *Names {
	names := &Names{
		n: make(map[string]bool, len(values)),
	}

	for _, v := range values {
		names.n[normalizeName(v)] = true
	}

	return names
}

// Len returns the count of service names in this set.
// If this Names is nil, this method returns zero (0).
func (n *Names) Len() int {
	if n == nil {
		return 0
	}

	return len(n.n)
}

// Has tests if this set contains the normalized version of a given service name.
// If this Names is nil, this method always returns false.
func (n *Names) Has(v string) bool {
	if n == nil {
		return false
	}

	_, has := n.n[normalizeName(v)]
	return has
}

// All provides iteration over each name in this set.
// If this Names is nil, this method does nothing.
func (n *Names) All(f func(string) bool) {
	if n != nil {
		for name := range n.n {
			if !f(name) {
				return
			}
		}
	}
}

// Merge returns a Names set that is the union of this Names
// with the given list. If list is empty, this method returns
// this Names as is. If this Names is nil or has zero length,
// NewNames is used to create a new Names with the given list.
func (n *Names) Merge(list ...string) *Names {
	switch {
	case len(list) == 0:
		return n

	case n == nil || n.Len() == 0:
		return NewNames(list...)

	default:
		merged := make(map[string]bool, n.Len()+len(list))
		for name := range n.n {
			merged[name] = true
		}

		for _, name := range list {
			merged[normalizeName(name)] = true
		}

		return &Names{n: merged}
	}
}

// Update tests if a list of names would represent an update to
// this set of names. If list represents a different set of names,
// accounting for duplicates, this method returns a distinct names
// instance and true. Otherwise, this method returns this names
// and false.
//
// Names are always immutable. This method does not modify the
// original Names set.
func (n *Names) Update(list ...string) (*Names, bool) {
	switch {
	case len(list) == 0:
		return n, false // no update

	case n == nil || n.Len() == 0:
		return NewNames(list...), true // the entire list is the update

	default:
		// have to account for any duplicates in the updated list
		updated := make(map[string]bool, len(list))
		sameNames := true
		for _, name := range list {
			name = normalizeName(name)
			updated[name] = true
			if sameNames {
				_, sameNames = n.n[name]
			}
		}

		if !sameNames || n.Len() != len(updated) {
			return &Names{n: updated}, true
		}

		return n, false
	}
}
