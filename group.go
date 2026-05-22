// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"iter"
	"net/netip"
	"slices"
)

// Server is a single endpoint of a service.
type Server struct {
	name         string
	originalName string

	a    []netip.Addr
	aaaa []netip.Addr
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

// A are all the addresses that came from DNS A records. These will
// be in sorted order.
func (s *Server) A() iter.Seq[netip.Addr] {
	return slices.Values(s.a)
}

// AAAA are all the addresses that came from DNS AAAA records. These
// will be in sorted order.
func (s *Server) AAAA() iter.Seq[netip.Addr] {
	return slices.Values(s.aaaa)
}

// Group is a single set of servers.
type Group struct {
	name string

	byOriginalName map[string]int
	servers        []Server
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

// newGroups returns a Groups with all the defined groups
// in the order of the given slice.
func newGroups(gdefs []GroupDefinition) *Groups {
	gps := &Groups{
		byName: make(map[string]int, len(gdefs)),
		all:    make([]Group, len(gdefs)),
	}

	for i, gdef := range gdefs {
		gps.byName[gdef.Name] = i
		gps.all[i] = Group{name: gdef.Name}
	}

	return gps
}

func (gps *Groups) Len() int {
	return len(gps.all)
}

func (gps *Groups) At(i int) *Group {
	return &gps.all[i]
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
