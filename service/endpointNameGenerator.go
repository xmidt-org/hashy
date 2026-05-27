package service

import (
	"fmt"
	"iter"
)

const (
	DefaultGeneratedEndpointNamePrefix = "hashy"
	DefaultGeneratedEndpointDomain     = "hashy.net"
)

// EndpointNameGenerator generates synthetic names for endpoints, based on their groupings.
type EndpointNameGenerator struct {
	// Prefix is the prefix for the generated names of servers. This must be a valid
	// DNS host label. If unset, defaults to DefaultGeneratedEndpointNamePrefix.
	Prefix string

	// Domain is the domain for all generated names of servers. If unset,
	// defaults to DefaultGeneratedEndpointDomain.
	Domain string
}

// GenerateNames generates synthetic names for all endpoints within the given Group.
// This method is idempotent.
func (gen EndpointNameGenerator) GenerateNames(g *Group) {
	prefix := gen.Prefix
	if len(prefix) == 0 {
		prefix = DefaultGeneratedEndpointNamePrefix
	}

	domain := gen.Domain
	if len(domain) == 0 {
		domain = DefaultGeneratedEndpointDomain
	}

	i := 0
	for endpoint := range g.Endpoints() {
		endpoint.name = fmt.Sprintf("%s-%d.%s.%s", prefix, i, g.name, domain)
		i++
	}
}

// GenerateAll generates names for all endpoints in each group in a sequence.
func (gen EndpointNameGenerator) GenerateAll(gseq iter.Seq[*Group]) {
	for g := range gseq {
		gen.GenerateNames(g)
	}
}
