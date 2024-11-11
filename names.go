package hashy

import (
	"bytes"
	"strings"
)

// normalizeName normalizes the given name.  Right now, all this
// function does is strings.ToLower.
func normalizeName(v string) string {
	return strings.ToLower(v)
}

// appendName adds a name to a Names set.  This function
// returns true if the Names was updated.
func appendName(n *Names, v string) bool {
	v = normalizeName(v)
	var exists bool
	if _, exists = n.m[v]; !exists {
		n.l += len(v)
		n.m[v] = true
	}

	return !exists
}

// appendNames adds several names to a Names set.
func appendNames(n *Names, list ...string) {
	for _, v := range list {
		appendName(n, v)
	}
}

// Names is an immutable set of service names.  Names are deduped
// and normalized. The zero value of this type is an empty set.
// Use NewNames to create a non-empty Names.
type Names struct {
	l int // the sum of the lengths of the names.  used as a hint for marshaling.
	m map[string]bool
}

// NewNames constructs an immutable Names set.
func NewNames(list ...string) (names Names) {
	names.m = make(map[string]bool, len(list))
	appendNames(&names, list...)
	return
}

// Len returns the count of service names in this set.
func (n Names) Len() int {
	return len(n.m)
}

// Has tests if this set contains the normalized version of a given service name.
func (n Names) Has(v string) bool {
	_, has := n.m[normalizeName(v)]
	return has
}

// All provides iteration over each name in this set.
func (n Names) All(f func(string) bool) {
	for name := range n.m {
		if !f(name) {
			return
		}
	}
}

// Merge produces a new Names instance that is the union of this Names and
// the given list. The result is deduped and normalized.
func (n Names) Merge(list ...string) (merged Names) {
	switch {
	case n.Len() == 0:
		merged = NewNames(list...)

	case len(list) == 0:
		merged = n

	default:
		merged = Names{
			l: n.l,
			m: make(map[string]bool, n.Len()+len(list)),
		}

		for existing := range n.m {
			merged.m[existing] = true
		}

		appendNames(&merged, list...)
	}

	return
}

// Update tests if a list of names would represent an update to
// this set of names. If list represents a different set of names,
// accounting for duplicates, this method returns a distinct names
// instance and true. Otherwise, this method returns this names
// and false.
func (n Names) Update(list ...string) (Names, bool) {
	switch {
	case len(list) == 0:
		return n, false // no update

	case n.Len() == 0:
		return NewNames(list...), true // the entire list is the update

	default:
		// have to account for any duplicates in the updated list
		updated := Names{
			m: make(map[string]bool, len(list)),
		}

		subset := true // whether the updated list is a subset of this Names
		for _, name := range list {
			if appendName(&updated, name) {
				if subset {
					_, subset = n.m[name]
				}
			}
		}

		if !subset || n.Len() != updated.Len() {
			return updated, true
		}

		return n, false
	}
}

// MarshalJSON marshals this set of names as a JSON array.
func (n *Names) MarshalJSON() ([]byte, error) {
	if n == nil {
		return []byte{'[', ']'}, nil
	}

	var b bytes.Buffer
	b.Grow(
		n.l + 3*len(n.m) + 1, // computes the space necessary for ["name1","name2",...]
	)

	b.WriteRune('[')
	for name := range n.m {
		if b.Len() > 1 {
			b.WriteRune(',')
		}

		b.WriteRune('"')
		b.WriteString(name)
		b.WriteRune('"')
	}

	b.WriteRune(']')
	return b.Bytes(), nil
}

// String returns a string representation of this names set.
func (n *Names) String() string {
	if n == nil {
		return ""
	}

	var o strings.Builder
	o.Grow(n.l + len(n.m) - 1) // computes the space necessary for name1,name2,...
	for name := range n.m {
		if o.Len() > 0 {
			o.WriteRune(',')
		}

		o.WriteString(name)
	}

	return o.String()
}
