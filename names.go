package hashy

import (
	"strings"
)

// normalizeName normalizes the given name.  Right now, all this
// function does is strings.ToLower.
func normalizeName(v string) string {
	return strings.ToLower(v)
}

// emptyNames is a canonical empty Names set.
var emptyNames *Names = new(Names)

// Names is a set of normalized, unique names for hashing.
// A Names instance is immutable once created.
//
// The zero value of this type is an empty Names set.  Use
// NewNames to create a non-empty set.
type Names struct {
	n map[string]bool
}

// NewNames constructs an immutable Names set from the given
// slice of names.  Names are normalized, and duplicate names
// are ignored.
func NewNames(names ...string) *Names {
	if len(names) == 0 {
		return emptyNames
	}

	n := &Names{
		n: make(map[string]bool, len(names)),
	}

	for _, name := range names {
		n.n[normalizeName(name)] = true
	}

	return n
}

// Len returns the count of names in this set.
func (n *Names) Len() int {
	return len(n.n)
}

// Has tests if this Names set has the given name.
func (n *Names) Has(v string) bool {
	_, exists := n.n[normalizeName(v)]
	return exists
}

// Each visits each name in this set. The predicate can return
// false to exit early. The order in which names are visited is undefined.
func (n *Names) Each(f func(string) bool) {
	for name := range n.n {
		if !f(name) {
			return
		}
	}
}

// Merge produces the union of this Names instance with another set.
func (n *Names) Merge(more *Names) *Names {
	switch {
	case more.Len() == 0:
		return n

	case n.Len() == 0:
		return more

	default:
		mergedMap := make(map[string]bool, n.Len()+more.Len())
		for name := range n.Each {
			mergedMap[name] = true
		}

		for name := range more.Each {
			mergedMap[name] = true
		}

		return &Names{n: mergedMap}
	}
}

// Update compares a possibly distinct list of names with this Names.
// If the newNames list is the same, accounting for duplicates, this
// method returns this existing Names instance along with false.
// If the newNames list is different, a new Names instance is returned
// along with true.
func (n *Names) Update(list ...string) (*Names, bool) {
	if len(list) == 0 {
		return emptyNames, len(n.n) != 0
	}

	// have to account for any duplicates in the updated list
	updatedMap := make(map[string]bool, len(list))
	newName := false
	for _, name := range list {
		name = normalizeName(name)
		updatedMap[name] = true
		if !newName {
			if _, exists := n.n[name]; !exists {
				newName = true
			}
		}
	}

	if newName || len(n.n) != len(updatedMap) {
		return &Names{n: updatedMap}, true
	}

	return n, false
}
