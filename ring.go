package hashy

import (
	"bytes"
	"sort"
	"strconv"
	"unsafe"
)

const (
	DefaultRingVNodes int = 200
)

type RingOption interface {
	apply(*Ring) error
}

type ringOptionFunc func(*Ring) error

func (f ringOptionFunc) apply(r *Ring) error { return f(r) }

// WithHash sets the Hash algorithm used by the Ring. By default,
// Murmur3 is used.
func WithHash(h Hash) RingOption {
	return ringOptionFunc(func(r *Ring) error {
		r.hash = h
		return nil
	})
}

// WithVNodes sets the number of replica nodes within the ring.
// The default is DefaultRingVNodes.
func WithVNodes(v int) RingOption {
	return ringOptionFunc(func(r *Ring) error {
		r.vnodes = v
		return nil
	})
}

// WithNames adds names to this ring.  Each invocation of this
// option will produce a merged, deduped set of names to occupy
// the ring.
func WithNames(n *Names) RingOption {
	return ringOptionFunc(func(r *Ring) error {
		r.names = r.names.Merge(n)
		return nil
	})
}

type ringNode struct {
	token uint64
	name  string
}

type ringNodes []ringNode

func (rn ringNodes) Len() int {
	return len(rn)
}

func (rn ringNodes) Less(i, j int) bool {
	return rn[i].token < rn[j].token
}

func (rn ringNodes) Swap(i, j int) {
	rn[i], rn[j] = rn[j], rn[i]
}

// Ring provides a hash ring for names.  A Ring is immutable once
// created. This implementation follows github.com/billhathaway/consistentHash
// for backwards compatibility with xmidt.
type Ring struct {
	vnodes int
	hash   Hash
	names  *Names
	nodes  ringNodes
}

// initialize builds the ringNodes for this Ring.
func (r *Ring) initialize() {
	var (
		hash       = r.hash.New64()
		nameBuffer bytes.Buffer
		prefix     = make([]byte, 0, 32) // grab enough that reallocations are unlikely
	)

	// preallocate as much as possible
	r.nodes = make(ringNodes, 0, r.names.Len()*r.vnodes)
	nameBuffer.Grow(256)

	for name := range r.names.Each {
		nameBuffer.Reset()
		nameBuffer.WriteString(name) // convert name to bytes only once

		for increment := int64(0); increment < int64(r.vnodes); increment++ {
			// this keeps the same format as github.com/billhathaway/consistentHash:
			// {integer}={name}
			hash.Reset()
			prefix = strconv.AppendInt(prefix[:0], increment, 10)
			prefix = append(prefix, '=')
			hash.Write(prefix)
			hash.Write(nameBuffer.Bytes())

			r.nodes = append(r.nodes,
				ringNode{
					token: hash.Sum64(),
					name:  name,
				},
			)
		}
	}

	sort.Sort(r.nodes)
}

// Update checks to see if the given set of names is different from the one
// currently hashed. If there are no differences, this Ring is returned as is
// along with false. If a new Ring had to be created, that new Ring is returned
// along with true.
//
// The returned Ring will always have the same configuration, e.g. vnodes, as
// this Ring.
func (r *Ring) Update(list ...string) (*Ring, bool) {
	if updatedNames, updated := r.names.Update(list...); updated {
		updatedRing := &Ring{
			hash:   r.hash,
			vnodes: r.vnodes,
			names:  updatedNames,
		}

		updatedRing.initialize()
		return updatedRing, true
	}

	return r, false
}

// Get is like GetBytes, but avoids unnecessary allocation when
// hashing a string.
func (r *Ring) Get(v string) string {
	if len(v) == 0 {
		return ""
	}

	return r.GetBytes(
		unsafe.Slice(unsafe.StringData(v), len(v)),
	)
}

// GetBytes hashes the given value and returns the closest address on the ring.
func (r *Ring) GetBytes(v []byte) string {
	if r.nodes.Len() == 0 {
		return ""
	}

	target := r.hash.Sum64(v)
	index := sort.Search(
		r.nodes.Len(),
		func(i int) bool {
			return r.nodes[i].token >= target
		},
	)

	if index >= r.nodes.Len() {
		index = 0
	}

	return r.nodes[index].name
}

// NewRing produces a hash ring with the given options.
func NewRing(opts ...RingOption) (*Ring, error) {
	r := new(Ring)
	for _, o := range opts {
		if err := o.apply(r); err != nil {
			return nil, err
		}
	}

	if r.hash == nil {
		r.hash = Murmur3{}
	}

	if r.vnodes <= 0 {
		r.vnodes = DefaultRingVNodes
	}

	r.initialize()
	return r, nil
}
