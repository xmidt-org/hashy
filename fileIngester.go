// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"

	"codeberg.org/miekg/dns"
)

// members unique set of a comparable type.
type members[E comparable] map[E]struct{}

func (m members[E]) collect() (s []E) {
	if len(m) > 0 {
		s = make([]E, 0, len(m))
		for k := range m {
			s = append(s, k)
		}
	}

	return
}

// membership holds a mapping of a container name, such as a group or service, to
// its members.
type membership[E comparable] map[string]members[E]

func (m *membership[E]) add(container string, element E) {
	switch {
	case *m == nil:
		*m = membership[E]{
			container: members[E]{
				element: {},
			},
		}

	case (*m)[container] != nil:
		(*m)[container][element] = struct{}{}

	default:
		(*m)[container] = members[E]{
			element: {},
		}
	}
}

func (m *membership[E]) addSeveral(container string, elements []E) {
	for _, element := range elements {
		m.add(container, element)
	}
}

// rrCollector collects RR records in preparation for building a set of Groups.
type rrCollector struct {
	// discoveryDomain is the TXT record owner for group information
	discoveryDomain string

	// groups is a mapping between group and service name (SRV record)
	groups membership[string]

	// services is a mapping between service and server
	services membership[string]

	// a maps server IDs onto IPv4 addresses
	a membership[netip.Addr]

	// aaaa maps server IDs onto IPv6 addresses
	aaaa membership[netip.Addr]
}

func (rrc *rrCollector) add(rr dns.RR) error {
	if rr.Header().Class != dns.ClassINET {
		// skip anything that isn't IN
		return nil
	}

	switch record := rr.(type) {
	case *dns.TXT:
		if record.Hdr.Name == rrc.discoveryDomain {
			for _, txt := range record.Txt {
				if gdef, err := ParseGroupDefinition(txt); err == nil {
					rrc.groups.addSeveral(gdef.Name, gdef.Services)
				} else {
					return err
				}
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

func (rrc *rrCollector) createGroups() (gps Groups) {
	gps.all = make([]Group, 0, len(rrc.groups))
	for groupName, serviceNames := range rrc.groups {
		g := Group{
			name: groupName,
		}

		for serviceName := range serviceNames {
			for serverID := range rrc.services[serviceName] {
				s := Server{
					id:          serverID,
					serviceName: serviceName,
					a:           rrc.a[serverID].collect(),
					aaaa:        rrc.aaaa[serverID].collect(),
				}

				if len(s.a) > 0 || len(s.aaaa) > 0 {
					// only add servers if addresses were defined
					g.servers = append(g.servers, s)
				}
			}
		}

		// if we couldn't determine any servers for this group, skip it
		if len(g.servers) == 0 {
			continue
		}

		gps.all = append(gps.all, g)
	}

	normalizeGroups(gps.all)
	return
}

type FileIngester struct {
	globs           []string
	origin          string
	defaultTTL      uint32
	discoveryDomain string

	lock      sync.RWMutex
	listeners []IngestListener
}

func (fi *FileIngester) AddIngestListener(l IngestListener) {
	fi.lock.Lock()
	fi.listeners = append(fi.listeners, l)
	fi.lock.Unlock()
}

func (fi *FileIngester) RemoveIngestListener(l IngestListener) {
	fi.lock.Lock()
	defer fi.lock.Unlock()

	for i := 0; i < len(fi.listeners); i++ {
		if fi.listeners[i] == l {
			fi.listeners = slices.Delete(fi.listeners, i, i+1)
			return
		}
	}
}

func (fi *FileIngester) ingestFile(ctx context.Context, rrc *rrCollector, path string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()
	zp := dns.NewZoneParser(file, fi.origin, path)
	zp.IncludeFS = os.DirFS(filepath.Dir(path)) // includes are relative to the location of the file
	zp.SetDefaultTTL(fi.defaultTTL)

	for rr, err := range zp.RRs() {
		if err != nil {
			return err
		}

		if err := rrc.add(rr); err != nil {
			return err
		}
	}

	return nil
}

func (fi *FileIngester) dispatchIngestEvent(event IngestEvent) {
	fi.lock.RLock()
	defer fi.lock.RUnlock()
	for _, l := range fi.listeners {
		l.OnIngest(event)
	}
}

func (fi *FileIngester) Ingest(ctx context.Context) {
	var event IngestEvent
	rrc := rrCollector{
		discoveryDomain: fi.discoveryDomain,
	}

	var paths []string
	for _, glob := range fi.globs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			event.Err = err
			break
		}

		// sort within each glob for a consistent processing order
		sort.Strings(matches)
		paths = append(paths, matches...)
	}

	if event.Err == nil {
		for _, path := range paths {
			if err := fi.ingestFile(ctx, &rrc, path); err != nil {
				event.Err = err
			}
		}
	}

	if event.Err == nil {
		event.Groups = rrc.createGroups()
	}

	fi.dispatchIngestEvent(event)
}
