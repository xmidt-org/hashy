package hashy

import (
	"hash"
	"sort"
	"strconv"
	"unsafe"
)

type ringNode struct {
	token   uint64
	service Service
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

func newRing(vnodes int, hash hash.Hash64, services Services) (r ring) {
	prefix := make([]byte, 0, 32) // grab enough that reallocations are unlikely
	r = make(ring, 0, services.Len()*vnodes)

	for service := range services.All {
		name := service.Name()
		nameBytes := unsafe.Slice(unsafe.StringData(name), len(name))
		for increment := int64(0); increment < int64(vnodes); increment++ {
			// this keeps the same format as github.com/billhathaway/consistentHash:
			// {integer}={name}
			hash.Reset()
			prefix = strconv.AppendInt(prefix[:0], increment, 10)
			prefix = append(prefix, '=')
			hash.Write(prefix)
			hash.Write(nameBytes)

			r = append(r, ringNode{
				token:   hash.Sum64(),
				service: service,
			})
		}
	}

	sort.Sort(r)
	return
}

// get returns the service on the ring whose token is the closest
// to the given target value.
func (r ring) get(target uint64) Service {
	index := sort.Search(
		r.Len(),
		func(i int) bool {
			return r[i].token >= target
		},
	)

	if index >= r.Len() {
		index = 0
	}

	return r[index].service
}
