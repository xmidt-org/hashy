package hashy

import (
	"fmt"
	"iter"
)

const (
	DefaultGeneratedServerNamePrefix = "hashy"
	DefaultGeneratedServerDomain     = "hashy.net"
)

// ServerNameGenerator generates synthetic names for servers, based on their groupings.
type ServerNameGenerator struct {
	// prefix is the prefix for the generated names of servers.
	prefix string

	// domain is the domain for all generated names of servers.
	domain string
}

// NewServerNameGenerator creates a new ServerNameGenerator.
//
// The prefix is the part prepended to generated names, followed by a hyphen.
// If blank, DefaultGeneratedServerNamePrefix is used.
//
// The domain is the zone domain that hashy serves. If blank, DefaultGeneratedServerDomain
// is used.
func NewServerNameGenerator(prefix, domain string) *ServerNameGenerator {
	return &ServerNameGenerator{
		prefix: prefix,
		domain: domain,
	}
}

func (gen *ServerNameGenerator) GenerateNames(groups iter.Seq[*Group]) {
	for g := range groups {
		i := 0
		for server := range g.Servers() {
			server.name = fmt.Sprintf("%s-%d.%s.%s", gen.prefix, i, g.name, gen.domain)
			i++
		}
	}
}
