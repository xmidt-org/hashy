// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"sort"
	"sync"
	"unsafe"
)

const (
	// DefaultVNodes is used for a Locator when no vnodes value is set.
	DefaultVNodes int = 200
)

// Group represents the name of a logical group of services.
// Service Locators can return services from multiple groups.
type Group string

type locatorNode struct {
	group Group
	names Names
	ring  ring
}

type locatorNodes []locatorNode

func (ln locatorNodes) Len() int {
	return len(ln)
}

func (ln locatorNodes) Less(i, j int) bool {
	return ln[i].group < ln[j].group
}

func (ln locatorNodes) Swap(i, j int) {
	ln[i], ln[j] = ln[j], ln[i]
}

// sort sorts this set of nodes by Group
func (ln locatorNodes) sort() {
	sort.Sort(ln)
}

// find locates the node associated with the given group and its
// index, if that group exists. If the group wasn't found, this method
// always returns nil and the length of the slice.
func (ln locatorNodes) find(g Group) (*locatorNode, int) {
	i := sort.Search(
		len(ln),
		func(i int) bool { return ln[i].group >= g },
	)

	if i < ln.Len() && ln[i].group == g {
		return &ln[i], i
	}

	return nil, ln.Len()
}

// append adds more names to a group, or creates the group if it doesn't
// exist. If the more slice is empty, this method does nothing.
func (ln locatorNodes) append(g Group, more ...string) locatorNodes {
	if len(more) > 0 {
		if existing, _ := ln.find(g); existing != nil {
			existing.names = existing.names.Merge(more...)
		} else {
			ln = append(ln, locatorNode{
				group: g,
				names: NewNames(more...),
			})

			ln.sort()
		}
	}

	return ln
}

// remove deletes a group and its names from this slice, if
// that group exists.
func (ln locatorNodes) remove(g Group) locatorNodes {
	if existing, index := ln.find(g); existing != nil {
		last := ln.Len() - 1
		ln[index], ln[last] = ln[last], locatorNode{}
		ln = ln[:last]
		ln.sort()
	}

	return ln
}

// insert puts or replaces a new node into this slice.
func (ln locatorNodes) insert(newNode locatorNode) locatorNodes {
	if _, index := ln.find(newNode.group); index < ln.Len() {
		ln[index] = newNode
	} else {
		ln = append(ln, newNode)
		ln.sort()
	}

	return ln
}

// Service is a located service.
type Service struct {
	// Group is any group associated with this service instance.
	Group Group

	// Name is the service name, which is normally a host name.
	Name string
}

// LocatorOption represents a configurable option for a service Locator.
type LocatorOption interface {
	apply(*locator) error
}

type locatorOptionFunc func(*locator) error

func (f locatorOptionFunc) apply(l *locator) error { return f(l) }

// WithVNodes configures the number of vnodes used per group by the locator.
// The default is DefaultVNodes. To effectively disable consistent hashing,
// set the number of vnodes to 1.
func WithVNodes(vnodes int) LocatorOption {
	return locatorOptionFunc(func(l *locator) error {
		l.vnodes = vnodes
		return nil
	})
}

// WithHash sets the hash algorithm used by the locator. The default
// used is Murmur3.
func WithHash(hash Hash) LocatorOption {
	return locatorOptionFunc(func(l *locator) error {
		l.hash = hash
		return nil
	})
}

// WithGroup initializes service names in a group for the locator. Multiple
// uses of this option are cumulative.
func WithGroup(g Group, more ...string) LocatorOption {
	return locatorOptionFunc(func(l *locator) error {
		l.nodes = l.nodes.append(g, more...)
		return nil
	})
}

// Locator is a service locator.
type Locator interface {
	// Find locates services associated with the given value.
	Find([]byte) []Service

	// FindGroup locates services associated with the given value,
	// but only within the specified group. If the given group does
	// not exist, this method returns an empty slice.
	FindGroup(Group, []byte) []Service

	// Remove atomically removes a group and all its service names
	// from this Locator.  If no such group exists, this method
	// does nothing.
	Remove(Group)

	// Update atomically updates the set of service names associated
	// with the given group.
	Update(Group, ...string)
}

// locator is the internal Locator implementation.
type locator struct {
	vnodes int
	hash   Hash

	lock  sync.RWMutex
	nodes locatorNodes
}

// initialize establishes defaults and builds any necessary hash rings for
// initial groups.
func (l *locator) initialize() {
	if l.vnodes < 1 {
		l.vnodes = DefaultVNodes
	}

	if l.hash == nil {
		l.hash = Murmur3{}
	}

	for _, ln := range l.nodes {
		ln.ring = l.newRing(ln.names)
	}
}

// newRing creates a hash ring using this locator's configuration
// along with the given service names.  This method does not
// require execution under the lock.
func (l *locator) newRing(names Names) ring {
	return newRing(l.vnodes, l.hash.New64(), names)
}

func (l *locator) Find(v []byte) (services []Service) {
	l.lock.RLock()

	if len(l.nodes) > 0 {
		services = make([]Service, 0, len(l.nodes))
		target := l.hash.Sum64(v)

		for _, ln := range l.nodes {
			services = append(services, Service{
				Group: ln.group,
				Name:  ln.ring.get(target),
			})
		}
	}

	l.lock.RUnlock()
	return
}

func (l *locator) FindGroup(g Group, v []byte) (services []Service) {
	l.lock.RLock()

	if node, _ := l.nodes.find(g); node != nil {
		services = []Service{
			Service{
				Group: node.group,
				Name:  node.ring.get(l.hash.Sum64(v)),
			},
		}
	}

	l.lock.RUnlock()
	return
}

func (l *locator) Remove(g Group) {
	l.lock.Lock()
	l.nodes = l.nodes.remove(g)
	l.lock.Unlock()
}

func (l *locator) Update(g Group, list ...string) {
	var updatedNames Names
	needsUpdate := true

	l.lock.RLock()

	existing, _ := l.nodes.find(g)
	if existing != nil {
		updatedNames, needsUpdate = existing.names.Update(list...)
	} else {
		updatedNames = NewNames(list...)
		needsUpdate = true // this is a new group
	}

	l.lock.RUnlock()

	if needsUpdate {
		// compute the new hash ring outside any lock
		newNode := locatorNode{
			group: g,
			names: updatedNames,
			ring:  l.newRing(updatedNames),
		}

		l.lock.Lock()

		// another goroutine may have barged and updated the same group.
		// so, we need to search again for that group's node, which is
		// what insert does.
		l.nodes = l.nodes.insert(newNode)

		l.lock.Unlock()
	}
}

// NewLocator creates a hash-based service locator from a set of options.
func NewLocator(opts ...LocatorOption) (Locator, error) {
	l := new(locator)
	for _, o := range opts {
		if err := o.apply(l); err != nil {
			return nil, err
		}
	}

	l.initialize()
	return l, nil
}

// Find uses the given locator to find services by a given string
// value. No additional memory allocation is performed, making this
// a better option that using []byte(value).
func Find(l Locator, v string) []Service {
	return l.Find(
		unsafe.Slice(unsafe.StringData(v), len(v)),
	)
}

// FindGroup uses the given Locator to find a service only within
// a certain group. No additional memory allocation is performed.
func FindGroup(l Locator, g Group, v string) []Service {
	return l.FindGroup(
		g,
		unsafe.Slice(unsafe.StringData(v), len(v)),
	)
}
