package hashy

import (
	"fmt"
	"iter"
)

type ServerNameGenerator struct {
	// serverNamePrefix is the prefix for the generated names of servers.
	serverNamePrefix string

	// generatedDomain is the domain for all generated names of servers.
	generatedDomain string
}

func (gen *ServerNameGenerator) GenerateNames(groups iter.Seq[*Group]) {
	for g := range groups {
		i := 0
		for server := range g.Servers() {
			server.name = fmt.Sprintf("%s-%d.%s.%s", gen.serverNamePrefix, i, g.name, gen.generatedDomain)
			i++
		}
	}
}
