package hashy

import (
	"fmt"
	"iter"
)

type ServerNameGenerator struct {
	// prefix is the prefix for the generated names of servers.
	prefix string

	// domain is the domain for all generated names of servers.
	domain string
}

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
