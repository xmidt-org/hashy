package hashy

import (
	"bytes"
	"hash"
	"sort"
	"strconv"
)

type ringNode struct {
	token uint64
	name  string
}

type ring []ringNode

func (r ring) Len() int {
	return len(r)
}

func (r ring) Less(i, j int) bool {
	return r[i].token < r[j].token
}

func (r ring) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func newRing(vnodes int, hash hash.Hash64, names *Names) (r ring) {
	var (
		nameBuffer bytes.Buffer
		prefix     = make([]byte, 0, 32) // grab enough that reallocations are unlikely
	)

	// preallocate as much as possible
	r = make(ring, 0, names.Len()*vnodes)
	nameBuffer.Grow(256)

	for name := range names.All {
		nameBuffer.Reset()
		nameBuffer.WriteString(name) // convert name to bytes only once

		for increment := int64(0); increment < int64(vnodes); increment++ {
			// this keeps the same format as github.com/billhathaway/consistentHash:
			// {integer}={name}
			hash.Reset()
			prefix = strconv.AppendInt(prefix[:0], increment, 10)
			prefix = append(prefix, '=')
			hash.Write(prefix)
			hash.Write(nameBuffer.Bytes())

			r = append(r, ringNode{
				token: hash.Sum64(),
				name:  name,
			})
		}
	}

	sort.Sort(r)
	return
}

// get returns the name on the ring whose token is the closest
// to the given target value.
func (r ring) get(target uint64) string {
	index := sort.Search(
		r.Len(),
		func(i int) bool {
			return r[i].token >= target
		},
	)

	if index >= r.Len() {
		index = 0
	}

	return r[index].name
}
