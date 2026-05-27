package service

import (
	"net/netip"
	"slices"

	"codeberg.org/miekg/dns"
)

const (
	DefaultDiscoveryDomain = "_hashy.discover"
)

// RRCollector collects DNS resource records in order to build a Groups. This type is basically
// the recipient of a stream of incoming RRs from an arbitrary source.
type RRCollector struct {
	// discoveryDomain is the DNS name that TXT records containing group information belong to
	discoveryDomain string

	nameGenerator EndpointNameGenerator

	// groups is group->service
	groups collector[string, string]

	// services is service->endpoint
	services collector[string, string]

	// a is endpoint->A record
	a collector[string, netip.Addr]

	// aaaa is endpoint->AAAA record
	aaaa collector[string, netip.Addr]
}

// AddRR adds an RR to this collector. Any RR that is not recognized is simply ignored.
func (rrc *RRCollector) AddRR(rr dns.RR) error {
	switch record := rr.(type) {
	case *dns.TXT:
		if record.Hdr.Name == rrc.discoveryDomain {
			for _, txt := range record.Txt {
				gdef, err := ParseGroupDefinition(txt)
				if err != nil {
					return err
				}

				rrc.groups.addSlice(gdef.Name, gdef.Services)
			}
		}

	case *dns.SRV:
		rrc.services.add(record.Hdr.Name, record.Target)

	case *dns.A:
		rrc.a.add(record.Hdr.Name, record.Addr)

	case *dns.AAAA:
		rrc.aaaa.add(record.Hdr.Name, record.Addr)
	}

	return nil
}

// Reset clears all collected records, but retains the underlying storage
// buffers for next time.
func (rrc *RRCollector) Reset() {
	rrc.groups.clear()
	rrc.services.clear()
	rrc.a.clear()
	rrc.aaaa.clear()
}

// allocateGroups creates one empty Group for each collected group. The returned
// Groups will be fully initialized with the groups in sorted order.
func (rrc *RRCollector) allocateGroups() *Groups {
	gps := &Groups{
		byName: make(map[string]int, rrc.groups.len()),
		all:    make([]Group, 0, rrc.groups.len()),
	}

	for name := range rrc.groups.sortedKeys() {
		gps.all = append(gps.all, Group{name: name})
		gps.byName[name] = len(gps.all) - 1
	}

	return gps
}

// buildGroup builds the Group from our collected data. The group's endpoints will be
// deduped and sorted. Each endpoint's addresses will also be deduped and sorted.
//
// The group must have previously been allocated via allocateGroups.
func (rrc *RRCollector) buildGroup(g *Group) {
	for serviceName := range rrc.groups.values(g.name) {
		// preallocated as we go ...
		g.endpoints = slices.Grow(g.endpoints, rrc.services.valuesLen(serviceName))

		for endpointName := range rrc.services.values(serviceName) {
			g.endpoints = append(g.endpoints, Endpoint{
				originalName: endpointName,
				a:            rrc.a.sortedValuesSlice(endpointName, compareAddr),
				aaaa:         rrc.a.sortedValuesSlice(endpointName, compareAddr),
			})
		}
	}

	slices.SortFunc(g.endpoints, compareEndpoints)
}

// Build constructs a Groups from the collected DNS RRs. After this method returns,
// this builder will be Reset.
func (rrc *RRCollector) Build() *Groups {
	gps := rrc.allocateGroups()
	for g := range gps.All() {
		rrc.buildGroup(g)
	}

	rrc.nameGenerator.GenerateAll(gps.All())
	rrc.Reset()
	return gps
}
