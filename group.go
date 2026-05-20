// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"iter"
	"net/netip"
	"slices"
	"strings"
)

// Addresses is a collection of sortable addresses.
type Addresses []netip.Addr

// Append adds more addresses, and
func (a Addresses) Append(more ...netip.Addr) Addresses {
	return append(a, more...)
}

// Sort sorts this set of Addresses in ascending order.
func (a Addresses) Sort() {
	slices.SortFunc(
		a,
		func(a1, a2 netip.Addr) int {
			return a1.Compare(a2)
		},
	)
}

// Server is a single endpoint of a service.
type Server struct {
	name         string
	originalName string

	a    Addresses
	aaaa Addresses
}

// Name is the synthetic, generated name for this server.
func (s *Server) Name() string {
	return s.name
}

// OriginalName is the name as it appeared in the source DNS records.
// This will likely be served from a different domain.
func (s *Server) OriginalName() string {
	return s.originalName
}

// A are all the addresses that came from DNS A records.
func (s *Server) A() iter.Seq[netip.Addr] {
	return slices.Values(s.a)
}

// AAAA are all the addresses that came from DNS AAAA records.
func (s *Server) AAAA() iter.Seq[netip.Addr] {
	return slices.Values(s.aaaa)
}

// normalize ensures that the address are sorted lexicographically.
// This is mainly so that any checksum over one or more servers is consistent.
func (s *Server) normalize() {
	s.a.Sort()
	s.aaaa.Sort()
}

// Group is a single set of servers.
type Group struct {
	name string

	byOriginalName map[string]int
	servers        []Server
}

// normalize will delete any servers with no addresses, normalize the remaining
// servers, then sort the remaining servers by originalName.
func (g *Group) normalize() {
	g.servers = slices.DeleteFunc(
		g.servers,
		func(s Server) bool {
			return len(s.a) == 0 && len(s.aaaa) == 0
		},
	)

	for i := range len(g.servers) {
		g.servers[i].normalize()
	}

	slices.SortFunc(
		g.servers,
		func(s1, s2 Server) int {
			return strings.Compare(s1.originalName, s2.originalName)
		},
	)

	g.byOriginalName = make(map[string]int, len(g.servers))
	for i, s := range g.servers {
		g.byOriginalName[s.originalName] = i
	}
}

func (g *Group) getOrAdd(originalServerName string) *Server {
	if pos, existing := g.byOriginalName[originalServerName]; existing {
		return &g.servers[pos]
	}

	newPos := len(g.servers) - 1
	g.servers = append(g.servers, Server{
		originalName: originalServerName,
	})

	g.byOriginalName[originalServerName] = newPos
	return &g.servers[newPos]
}

func (g *Group) Len() int {
	return len(g.servers)
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Servers() iter.Seq[*Server] {
	return func(yield func(*Server) bool) {
		for i := range len(g.servers) {
			if !yield(&g.servers[i]) {
				return
			}
		}
	}
}

// Groups is an immutable collection of Group.
type Groups struct {
	byName map[string]int
	all    []Group
}

// newGroups returns an initialized, empty Groups.
func newGroups() *Groups {
	return &Groups{
		byName: make(map[string]int),
	}
}

// normalize deletes any Group that has no servers, normalizes the remaining groups,
// then sorts the remaining groups by group name.
func (gps *Groups) normalize() {
	gps.all = slices.DeleteFunc(
		gps.all,
		func(g Group) bool {
			return len(g.servers) == 0
		},
	)

	for i := range len(gps.all) {
		gps.all[i].normalize()
	}

	slices.SortFunc(
		gps.all,
		func(g1, g2 Group) int {
			return strings.Compare(g1.name, g2.name)
		},
	)

	gps.byName = make(map[string]int, len(gps.all))
	for i, g := range gps.all {
		gps.byName[g.name] = i
	}
}

func (gps *Groups) getOrAdd(groupName string) *Group {
	if pos, existing := gps.byName[groupName]; existing {
		return &gps.all[pos]
	}

	newPos := len(gps.all) - 1
	gps.all = append(gps.all, Group{
		name:           groupName,
		byOriginalName: make(map[string]int),
	})

	gps.byName[groupName] = newPos
	return &gps.all[newPos]
}

func (gps *Groups) Len() int {
	return len(gps.all)
}

func (gps *Groups) Get(groupName string) *Group {
	if pos, existing := gps.byName[groupName]; existing {
		return &gps.all[pos]
	}

	return nil
}

func (gps *Groups) All() iter.Seq[*Group] {
	return func(yield func(*Group) bool) {
		for i := range len(gps.all) {
			if !yield(&gps.all[i]) {
				return
			}
		}
	}
}
