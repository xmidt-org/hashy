// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"iter"
	"slices"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
)

// Group is a single group of servers.
type Group struct {
	name      string
	services  []string
	endpoints []Endpoint
}

func (g *Group) Len() int {
	return len(g.endpoints)
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Services() iter.Seq[string] {
	return slices.Values(g.services)
}

func (g *Group) Endpoints() iter.Seq[*Endpoint] {
	return func(yield func(*Endpoint) bool) {
		for i := range len(g.endpoints) {
			if !yield(&g.endpoints[i]) {
				return
			}
		}
	}
}

// LenRRs returns the number of RRs this group will produce of the given type.
func (g *Group) LenRRs(rrType uint16) (n int) {
	if rrType == dns.TypeTXT {
		n = len(g.services) + len(g.endpoints)
	}

	return
}

// RRs produces a sequence of RR records for this group. If rrType is anything
// but TypeTXT, this method returns an empty sequence.
//
// Each returned RR has a header with only the Class set to ClassINET.
func (g *Group) RRs(rrType uint16) iter.Seq2[*Group, dns.RR] {
	if rrType != dns.TypeTXT {
		return emptyRRs[*Group]
	}

	return func(yield func(*Group, dns.RR) bool) {
		var buffer strings.Builder
		for _, serviceName := range g.services {
			buffer.Reset()
			buffer.WriteString(g.name)
			buffer.WriteByte(' ')
			buffer.WriteString(serviceName)

			rr := &dns.TXT{
				Hdr: dns.Header{
					Class: dns.ClassINET,
				},
				TXT: rdata.TXT{
					Txt: []string{buffer.String()},
				},
			}

			if !yield(g, rr) {
				return
			}
		}

		for _, endpoint := range g.endpoints {
			buffer.Reset()
			buffer.WriteString(g.name)
			buffer.WriteByte(' ')
			buffer.WriteString(endpoint.OriginalName())

			rr := &dns.TXT{
				Hdr: dns.Header{
					Class: dns.ClassINET,
				},
				TXT: rdata.TXT{
					Txt: []string{buffer.String()},
				},
			}

			if !yield(g, rr) {
				return
			}
		}
	}
}

// Groups is an immutable collection of Group instances.
type Groups struct {
	byName map[string]int
	all    []Group
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

func (gps *Groups) LenRRs(rrType uint16) (n int) {
	for _, g := range gps.all {
		n += g.LenRRs(rrType)
	}

	return
}

func (gps *Groups) RRs(rrType uint16) iter.Seq2[*Group, dns.RR] {
	return func(yield func(*Group, dns.RR) bool) {
		for _, group := range gps.all {
			for g, rr := range group.RRs(rrType) {
				if !yield(g, rr) {
					return
				}
			}
		}
	}
}
