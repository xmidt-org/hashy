package hashy

import (
	"iter"
	"maps"
	"net/netip"

	"codeberg.org/miekg/dns"
)

type collectedGroupEntry struct {
	groupName   string
	serviceName string
}

// collectedGroups holds the group: service tuples we've discovered.
type collectedGroups map[collectedGroupEntry]struct{}

func (cg collectedGroups) add(groupName, serviceName string) {
	cg[collectedGroupEntry{
		groupName:   groupName,
		serviceName: serviceName,
	}] = struct{}{}
}

func (cg collectedGroups) addDefinition(gdef GroupDefinition) {
	for _, service := range gdef.Services {
		cg.add(gdef.Name, service)
	}
}

func (cg collectedGroups) all() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for ge := range cg {
			if !yield(ge.groupName, ge.serviceName) {
				return
			}
		}
	}
}

// collectedServices holds all the service: server tuples. The servers are deduped.
type collectedServices map[string]map[string]struct{}

func (cs collectedServices) add(service, server string) {
	if members, existing := cs[service]; existing {
		members[server] = struct{}{}
	} else {
		cs[service] = map[string]struct{}{server: struct{}{}}
	}
}

func (cs collectedServices) servers(serviceName string) iter.Seq[string] {
	return maps.Keys(cs[serviceName])
}

type collectedAddressEntry struct {
	a    bool
	addr netip.Addr
}

// collectedServers holds all the server addresses. The addresses are deduped.
type collectedServers map[string]map[collectedAddressEntry]struct{}

func (cs collectedServers) addEntry(serverName string, entry collectedAddressEntry) {
	if members, existing := cs[serverName]; existing {
		members[entry] = struct{}{}
	} else {
		cs[serverName] = map[collectedAddressEntry]struct{}{entry: struct{}{}}
	}
}

func (cs collectedServers) addA(serverName string, addr netip.Addr) {
	cs.addEntry(
		serverName,
		collectedAddressEntry{
			a:    true,
			addr: addr,
		},
	)
}

func (cs collectedServers) addAAAA(serverName string, addr netip.Addr) {
	cs.addEntry(
		serverName,
		collectedAddressEntry{
			a:    false,
			addr: addr,
		},
	)
}

// addresses provides a sequence of (flag, address) tuples for a given server.
// The flag is true if the address came from an A record, false if it came from an AAAA record.
func (cs collectedServers) addresses(serverName string) iter.Seq2[bool, netip.Addr] {
	return func(yield func(bool, netip.Addr) bool) {
		for entry := range cs[serverName] {
			if !yield(entry.a, entry.addr) {
				return
			}
		}
	}
}

type RRCollector struct {
	// discoveryDomain is the DNS name that TXT records containing group information belong to
	discoveryDomain string

	nameGenerator *ServerNameGenerator

	groups   collectedGroups
	services collectedServices
	servers  collectedServers
}

func (rrc *RRCollector) addTXT(rr *dns.TXT) error {
	if rr.Hdr.Name != rrc.discoveryDomain {
		return nil
	}

	for _, txt := range rr.Txt {
		gdef, err := ParseGroupDefinition(txt)
		if err != nil {
			return err
		}

		rrc.groups.addDefinition(gdef)
	}

	return nil
}

func (rrc *RRCollector) addSRV(rr *dns.SRV) {
	if members, existing := rrc.services[rr.Hdr.Name]; existing {
		members[rr.Target] = struct{}{}
	} else {
		rrc.services[rr.Hdr.Name] = map[string]struct{}{rr.Target: struct{}{}}
	}
}

func (rrc *RRCollector) addA(rr *dns.A) {
	rrc.servers.addA(rr.Hdr.Name, rr.Addr)
}

func (rrc *RRCollector) addAAAA(rr *dns.AAAA) {
	rrc.servers.addAAAA(rr.Hdr.Name, rr.Addr)
}

func (rrc *RRCollector) Add(rr dns.RR) error {
	switch record := rr.(type) {
	case *dns.TXT:
		if err := rrc.addTXT(record); err != nil {
			return err
		}

	case *dns.SRV:
		rrc.addSRV(record)

	case *dns.A:
		rrc.addA(record)

	case *dns.AAAA:
		rrc.addAAAA(record)
	}

	return nil
}

func (rrc *RRCollector) Build() (gps *Groups) {
	gps = newGroups()

	for groupName, serviceName := range rrc.groups.all() {
		g := gps.getOrAdd(groupName)

		for serverName := range rrc.services.servers(serviceName) {
			for isA, addr := range rrc.servers.addresses(serverName) {
				s := g.getOrAdd(serverName)
				if isA {
					s.a = s.a.Append(addr)
				} else {
					s.aaaa = s.aaaa.Append(addr)
				}
			}
		}
	}

	gps.normalize()
	rrc.nameGenerator.GenerateNames(gps.All())

	clear(rrc.groups)
	clear(rrc.services)
	clear(rrc.servers)

	return
}
