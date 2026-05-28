package hashy

import "net/netip"

// CompareAddrs returns hashy's standard way of comparing addresses, which is
// to just use netip.Addr.Compare.
func CompareAddrs(a1, a2 netip.Addr) int {
	return a1.Compare(a2)
}
