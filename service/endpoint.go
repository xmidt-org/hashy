package service

import (
	"net/netip"

	"github.com/xmidt-org/hashy"
)

// Endpoint is a single endpoint of a service.
type Endpoint struct {
	originalName string

	ip4 hashy.Values[netip.Addr]
	ip6 hashy.Values[netip.Addr]
}

// OriginalName is the name as it appeared in the source DNS records.
// This will likely be served from a different domain.
func (s *Endpoint) OriginalName() string {
	return s.originalName
}
