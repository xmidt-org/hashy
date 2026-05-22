package hashy

import (
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

	nameGenerator *ServerNameGenerator

	groups   GroupDefinitionCollector
	services serviceCollector
	servers  serverDefinitionCollector
}

// AddRR adds an RR to this collector. Any RR that is not recognized is simply ignored.
func (rrc *RRCollector) AddRR(rr dns.RR) error {
	switch record := rr.(type) {
	case *dns.TXT:
		if record.Hdr.Name == rrc.discoveryDomain {
			if err := rrc.groups.AddText(record.Txt...); err != nil {
				return err
			}
		}

	case *dns.SRV:
		rrc.services.add(record.Hdr.Name, record.Target)

	case *dns.A:
		rrc.servers.addA(record.Hdr.Name, record.Addr)

	case *dns.AAAA:
		rrc.servers.addAAAA(record.Hdr.Name, record.Addr)
	}

	return nil
}

// Reset clears all collected records, but retains the underlying storage
// buffers for next time.
func (rrc *RRCollector) Reset() {
	rrc.groups.Clear()
	rrc.services.clear()
	rrc.servers.clear()
}

// Build constructs a Groups from the collected DNS RRs. After this method returns,
// this builder will be Reset.
func (rrc *RRCollector) Build() *Groups {
	gdefs := rrc.groups.Collect()
	gps := newGroups(gdefs)

	for i, gdef := range gdefs {
		group := gps.At(i)
		targets := rrc.services.targets(gdef.Services...)
		group.servers = make([]Server, 0, len(targets))

		for _, target := range targets {
			if s, ok := rrc.servers.newServer(target); ok {
				group.servers = append(group.servers, s)
			}
		}
	}

	rrc.nameGenerator.GenerateNames(gps.All())
	rrc.Reset()
	return gps
}
