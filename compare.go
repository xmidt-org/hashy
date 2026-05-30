// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import "net/netip"

// CompareAddrs returns hashy's standard way of comparing addresses, which is
// to just use netip.Addr.Compare.
func CompareAddrs(a1, a2 netip.Addr) int {
	return a1.Compare(a2)
}
