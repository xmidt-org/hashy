package service

import (
	"maps"
	"net/netip"
	"slices"
)

// serverDefinition what we know about a server.
type serverDefinition struct {
	originalName string
	a            map[netip.Addr]struct{}
	aaaa         map[netip.Addr]struct{}
}

// newServer creates a Server from this definition. The A and AAAA records
// will be deduped and sorted.
func (sdef serverDefinition) newServer() (s Endpoint) {
	s.originalName = sdef.originalName

	if len(sdef.a) > 0 {
		s.a = make([]netip.Addr, 0, len(sdef.a))
		s.a = slices.AppendSeq(s.a, maps.Keys(sdef.a))
		slices.SortFunc(
			s.a,
			func(left, right netip.Addr) int {
				return left.Compare(right)
			},
		)
	}

	if len(sdef.aaaa) > 0 {
		s.aaaa = make([]netip.Addr, 0, len(sdef.aaaa))
		s.aaaa = slices.AppendSeq(s.a, maps.Keys(sdef.aaaa))
		slices.SortFunc(
			s.aaaa,
			func(left, right netip.Addr) int {
				return left.Compare(right)
			},
		)
	}

	return
}

func (sdef *serverDefinition) addA(addr netip.Addr) {
	if sdef.a == nil {
		sdef.a = make(map[netip.Addr]struct{})
	}

	sdef.a[addr] = struct{}{}
}

func (sdef *serverDefinition) addAAAA(addr netip.Addr) {
	if sdef.aaaa == nil {
		sdef.aaaa = make(map[netip.Addr]struct{})
	}

	sdef.aaaa[addr] = struct{}{}
}

// serverDefinitionCollector collects information about servers and provide deduping and sorting
// of address records.
type serverDefinitionCollector map[string]*serverDefinition

func (sdc serverDefinitionCollector) clear() {
	clear(sdc)
}

func (sdc *serverDefinitionCollector) addA(serverName string, addrs ...netip.Addr) {
	if *sdc == nil {
		*sdc = make(serverDefinitionCollector)
	}

	sdef := (*sdc)[serverName]
	if sdef == nil {
		sdef = new(serverDefinition)
		sdef.originalName = serverName
		(*sdc)[serverName] = sdef
	}

	for _, addr := range addrs {
		sdef.a[addr] = struct{}{}
	}
}

func (sdc *serverDefinitionCollector) addAAAA(serverName string, addrs ...netip.Addr) {
	if *sdc == nil {
		*sdc = make(serverDefinitionCollector)
	}

	sdef := (*sdc)[serverName]
	if sdef == nil {
		sdef = new(serverDefinition)
		sdef.originalName = serverName
		(*sdc)[serverName] = sdef
	}

	for _, addr := range addrs {
		sdef.aaaa[addr] = struct{}{}
	}
}

// newServer creates a Server from the given name. The A and AAAA records
// will be deduped and sorted.
//
// If no such server exists in this collector, a blank Server is returned along with false.
func (sdc serverDefinitionCollector) newServer(serverName string) (Endpoint, bool) {
	if sdef := sdc[serverName]; sdef != nil {
		return sdef.newServer(), true
	}

	return Endpoint{originalName: serverName}, false
}
