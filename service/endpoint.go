package service

import (
	"iter"
	"net/netip"
	"slices"
	"strings"
)

func compareAddr(a1, a2 netip.Addr) int {
	return a1.Compare(a2)
}

func compareEndpoints(e1, e2 Endpoint) int {
	return strings.Compare(e1.originalName, e2.originalName)
}

// Endpoint is a single endpoint of a service.
type Endpoint struct {
	name         string
	originalName string

	a    []netip.Addr
	aaaa []netip.Addr
}

// Name is the synthetic, generated name for this endpoint.
func (s *Endpoint) Name() string {
	return s.name
}

// OriginalName is the name as it appeared in the source DNS records.
// This will likely be served from a different domain.
func (s *Endpoint) OriginalName() string {
	return s.originalName
}

// A are all the addresses that came from DNS A records. These will
// be in sorted order.
func (s *Endpoint) A() iter.Seq[netip.Addr] {
	return slices.Values(s.a)
}

// AAAA are all the addresses that came from DNS AAAA records. These
// will be in sorted order.
func (s *Endpoint) AAAA() iter.Seq[netip.Addr] {
	return slices.Values(s.aaaa)
}
