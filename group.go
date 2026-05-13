// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"net/netip"
	"slices"
	"strings"
)

func compareAddrs(a1, a2 netip.Addr) int {
	return a1.Compare(a2)
}

// Server is a single endpoint of a service.
type Server struct {
	id          string
	serviceName string

	a    []netip.Addr
	aaaa []netip.Addr
}

// normalize ensures that the address are sorted lexicographically.
// This is mainly so that any checksum over one or more servers is consistent.
func (s *Server) normalize() {
	slices.SortFunc(s.a, compareAddrs)
	slices.SortFunc(s.aaaa, compareAddrs)
}

func compareServers(s1, s2 Server) int {
	return strings.Compare(s1.id, s2.id)
}

// normalizeServers sorts a slice of servers lexicographically using service ID.
// This function also sorts the address records for each Server, but does not
// do any deduplication.
func normalizeServers(s []Server) {
	for i := 0; i < len(s); i++ {
		s[i].normalize()
	}

	slices.SortFunc(s, compareServers)
}

// Group is a single set of servers.
type Group struct {
	name    string
	servers []Server
}

func compareGroups(g1, g2 Group) int {
	return strings.Compare(g1.name, g2.name)
}

// normalizeGroups sorts a slice of groups lexicographically using group name.
// This also normalizes the servers within each group, but does not do any deduplication.
func normalizeGroups(gps []Group) {
	for i := 0; i < len(gps); i++ {
		normalizeServers(gps[i].servers)
	}

	slices.SortFunc(gps, compareGroups)
}

// Groups is an immutable collection of Group.
type Groups struct {
	all []Group
}
