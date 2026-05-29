package service

import (
	"net/netip"
	"slices"
	"sort"
	"strings"

	"codeberg.org/miekg/dns"
	"github.com/xmidt-org/hashy"
)

const (
	DefaultDiscoveryDomain = "_hashy.discover"
)

// groupCollector is a collector type for groups. As RRs are encountered that have
// group information, they are added to this map and the definitions are merged.
type groupCollector map[string]GroupDefinition

func (gc groupCollector) clear() {
	clear(gc)
}

func (gc *groupCollector) add(gdef GroupDefinition) {
	if *gc == nil {
		*gc = groupCollector{
			gdef.Name: gdef,
		}

		return
	}

	merged := (*gc)[gdef.Name]
	merged.Name = gdef.Name
	merged.Services = append(merged.Services, gdef.Services...)
	(*gc)[gdef.Name] = merged
}

// addTxt parses several group definitions and adds them to this collector.
// Any error in parsing is returned and will halt processing of subsequent texts.
func (gc *groupCollector) addTxt(values ...string) error {
	for _, txt := range values {
		gdef, err := ParseGroupDefinition(txt)
		if err != nil {
			return err
		}

		gc.add(gdef)
	}

	return nil
}

// sorted returns a sorted slice of all merged definitions in this set.
// Each definition's services will also be deduped and sorted.
func (gc groupCollector) sorted() (defs []GroupDefinition) {
	if len(gc) < 1 {
		return
	}

	defs = make([]GroupDefinition, 0, len(gc))
	var dedupeServices hashy.Deduper[string]
	for _, def := range gc {
		dedupeServices.Clear()
		dedupeServices.Add(def.Services...)
		def.Services = dedupeServices.AppendTo(def.Services[:0]) // reuse def.Services storage
		sort.Strings(def.Services)

		defs = append(defs, def)
	}

	slices.SortFunc(
		defs,
		func(def1, def2 GroupDefinition) int {
			return strings.Compare(def1.Name, def2.Name)
		},
	)

	return
}

// newGroups creates a new Groups from this collected set. The groups will be sorted
// by name and will have no endpoints. Each group's services will be deduped and sorted.
func (gc groupCollector) newGroups() *Groups {
	gps := &Groups{
		byName: make(map[string]int, len(gc)),
		all:    make([]Group, 0, len(gc)),
	}

	for _, gdef := range gc.sorted() {
		gps.all = append(gps.all, Group{
			name:     gdef.Name,
			services: gdef.Services,
		})
	}

	for i, g := range gps.all {
		gps.byName[g.name] = i
	}

	return gps
}

// servicesCollector collects service->target pairs.
type servicesCollector map[string]hashy.Values[string]

func (sc servicesCollector) clear() {
	clear(sc)
}

func (sc *servicesCollector) add(serviceName, targetName string) {
	if *sc == nil {
		*sc = servicesCollector{
			serviceName: hashy.Values[string]{targetName},
		}

		return
	}

	(*sc)[serviceName] = (*sc)[serviceName].Append(targetName)
}

// targetNames returns a slice of target names that belong to any of the supplied services.
// The returned slice is deduped and sorted.
func (sc servicesCollector) targetNames(serviceNames []string) []string {
	var dedupe hashy.Deduper[string]
	for _, serviceName := range serviceNames {
		dedupe.Add(sc[serviceName]...)
	}

	return dedupe.SortFunc(strings.Compare)
}

// endpointCollector collects information about endpoints (targets).
type endpointCollector map[string]Endpoint

func (ec endpointCollector) clear() {
	clear(ec)
}

func (ec *endpointCollector) addIP4(originalName string, addr netip.Addr) {
	if *ec == nil {
		*ec = endpointCollector{
			originalName: Endpoint{
				originalName: originalName,
				ip4:          hashy.Values[netip.Addr]{addr},
			},
		}

		return
	}

	edef := (*ec)[originalName]
	edef.originalName = originalName
	edef.ip4.Add(addr)
	(*ec)[originalName] = edef
}

func (ec *endpointCollector) addIP6(originalName string, addr netip.Addr) {
	if *ec == nil {
		*ec = endpointCollector{
			originalName: Endpoint{
				originalName: originalName,
				ip6:          hashy.Values[netip.Addr]{addr},
			},
		}

		return
	}

	edef := (*ec)[originalName]
	edef.originalName = originalName
	edef.ip6.Add(addr)
	(*ec)[originalName] = edef
}

// endpointsFor returns a slice of Endpoints corresponding to elements of a slice
// of target names. The output is 1-1 with the input names. No deduping or sorting is done
// by this method.
func (ec endpointCollector) endpointsFor(names []string) []Endpoint {
	endpoints := make([]Endpoint, 0, len(names))
	for _, n := range names {
		if endpoint, exists := ec[n]; exists {
			endpoint.ip4.Dedupe()
			endpoint.ip4.SortFunc(hashy.CompareAddrs)

			endpoint.ip6.Dedupe()
			endpoint.ip6.SortFunc(hashy.CompareAddrs)

			endpoints = append(endpoints, endpoint)
		}
	}

	return endpoints
}

// RRCollector collects DNS resource records in order to build a Groups. This type is basically
// the recipient of a stream of incoming RRs from an arbitrary source.
type RRCollector struct {
	// discoveryDomain is the DNS name that TXT records containing group information belong to
	discoveryDomain string

	// groups is group->service
	groups groupCollector

	// services is service->endpoint
	services servicesCollector

	// endpoints is target (server) -> Endpoint
	endpoints endpointCollector
}

// AddRR adds an RR to this collector. Any RR that is not recognized is simply ignored.
func (rrc *RRCollector) AddRR(rr dns.RR) error {
	switch record := rr.(type) {
	case *dns.TXT:
		if record.Hdr.Name == rrc.discoveryDomain {
			if err := rrc.groups.addTxt(record.Txt...); err != nil {
				return err
			}
		}

	case *dns.SRV:
		rrc.services.add(record.Hdr.Name, record.Target)

	case *dns.A:
		rrc.endpoints.addIP4(record.Hdr.Name, record.Addr)

	case *dns.AAAA:
		rrc.endpoints.addIP6(record.Hdr.Name, record.Addr)
	}

	return nil
}

// Reset clears all collected records, but retains the underlying storage
// buffers for next time.
func (rrc *RRCollector) Reset() {
	rrc.groups.clear()
	rrc.services.clear()
	rrc.endpoints.clear()
}

// Build constructs a Groups from the collected DNS RRs. After this method returns,
// this builder will be Reset.
func (rrc *RRCollector) Build() *Groups {
	gps := rrc.groups.newGroups()
	for g := range gps.All() {
		g.endpoints = rrc.endpoints.endpointsFor(
			rrc.services.targetNames(g.services),
		)
	}

	rrc.Reset()
	return gps
}
