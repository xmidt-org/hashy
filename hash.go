package hashy

import (
	"hash"

	"github.com/spaolacci/murmur3"
)

// Hash represents a hashing algorithm. This interface defines
// the behavior of a hash API expected by hashy.
type Hash interface {
	// New creates a 64-bit hash.
	New64() hash.Hash64

	// Sum64 returns a 64-bit hash value from the given set of bytes.
	Sum64([]byte) uint64
}

// Murmur3 is a Hash implementation that uses github.com/spaolacci/murmur3.
// This is the default Hash used by hashy.
type Murmur3 struct{}

func (m Murmur3) New64() hash.Hash64 {
	return murmur3.New64()
}

func (m Murmur3) Sum64(v []byte) uint64 {
	return murmur3.Sum64(v)
}
