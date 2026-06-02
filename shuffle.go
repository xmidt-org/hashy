package hashy

import (
	"math/rand/v2"

	"codeberg.org/miekg/dns"
)

// Shuffle does a random, in-place shuffle of a slice of RRs.
// If rrs is empty or only has 1 element, this function is a no-op.
//
// This function is useful when trying to make sure clients get a random
// set of records so they don't fixate on the first record in a set.
func Shuffle(rrs []dns.RR) {
	rand.Shuffle(len(rrs), func(i, j int) {
		rrs[i], rrs[j] = rrs[j], rrs[i]
	})
}
