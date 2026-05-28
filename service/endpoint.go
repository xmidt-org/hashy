package service

import (
	"net/netip"

	"github.com/xmidt-org/hashy"
)

// Endpoint is a single endpoint of a service.
type Endpoint struct {
	name         string
	originalName string

	ip4 hashy.Values[netip.Addr]
	ip6 hashy.Values[netip.Addr]
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
