// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"iter"
	"sync"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/rdata"
	"github.com/xmidt-org/medley"
	"github.com/xmidt-org/medley/consistent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LocatedEndpoints is a collection of endpoints that were discovered by a Locator.
type LocatedEndpoints []*Endpoint

// LenRRs returns the total count of all RRs of a given type in this set.
// This will be the number of tuples returned by the RRs() sequence.
func (le LocatedEndpoints) LenRRs(rrType uint16) (n int) {
	switch rrType {
	case dns.TypeA:
		for _, endpoint := range le {
			n += len(endpoint.ip4)
		}

	case dns.TypeAAAA:
		for _, endpoint := range le {
			n += len(endpoint.ip6)
		}
	}

	return
}

// RRs returns a sequence of (*Endpoint, dns.RR) tuples. Each RR is of the
// given type. Each RR's Class will be set to ClassINET, but will otherwise
// be uninitialized.
func (le LocatedEndpoints) RRs(rrType uint16) iter.Seq2[*Endpoint, dns.RR] {
	switch rrType {
	case dns.TypeA:
		return func(yield func(*Endpoint, dns.RR) bool) {
			for _, endpoint := range le {
				for _, addr := range endpoint.ip4 {
					rr := &dns.A{
						Hdr: dns.Header{
							Class: dns.ClassINET,
						},
						A: rdata.A{
							Addr: addr,
						},
					}

					if !yield(endpoint, rr) {
						return
					}
				}
			}
		}

	case dns.TypeAAAA:
		return func(yield func(*Endpoint, dns.RR) bool) {
			for _, endpoint := range le {
				for _, addr := range endpoint.ip6 {
					rr := &dns.AAAA{
						Hdr: dns.Header{
							Class: dns.ClassINET,
						},
						AAAA: rdata.AAAA{
							Addr: addr,
						},
					}

					if !yield(endpoint, rr) {
						return
					}
				}
			}
		}

	default:
		return emptyRRs[*Endpoint]
	}
}

// Locator is a service locator backed by one or more medley consistent hash Rings.
type Locator struct {
	logger  *zap.Logger
	builder *consistent.Builder[string, *Endpoint]

	lock        sync.RWMutex
	groups      *Groups
	ringsByName map[string]*consistent.Ring[*Endpoint]
	allRings    []*consistent.Ring[*Endpoint]
}

// rings produces a sequence of rings, optionally filtered by group.
// If no groups are passed, all rings are returned. If any groups are missing,
// no ring is pushed for that group name.
//
// No concurrency protection is provided by this method.  Callers must contend on the lock.
func (l *Locator) rings(groups []string) iter.Seq[*consistent.Ring[*Endpoint]] {
	return func(yield func(*consistent.Ring[*Endpoint]) bool) {
		if len(groups) > 0 {
			for _, groupName := range groups {
				if ring := l.ringsByName[groupName]; ring != nil {
					if !yield(ring) {
						return
					}
				}
			}
		} else {
			for _, ring := range l.allRings {
				if !yield(ring) {
					return
				}
			}
		}
	}
}

func (l *Locator) Find(object []byte, groups ...string) (results LocatedEndpoints) {
	defer l.lock.RUnlock()
	l.lock.RLock()

	results = make(LocatedEndpoints, 0, len(l.allRings)) // worst case
	for ring := range l.rings(groups) {
		results = append(results, ring.Nearest(object))
	}

	return
}

func (l *Locator) FindString(object string, groups ...string) (results LocatedEndpoints) {
	defer l.lock.RUnlock()
	l.lock.RLock()

	results = make(LocatedEndpoints, 0, len(l.allRings)) // worst case
	for ring := range l.rings(groups) {
		results = append(results, ring.NearestString(object))
	}

	return
}

func (l *Locator) Groups() (gps *Groups) {
	l.lock.RLock()
	gps = l.groups
	l.lock.RUnlock()

	return
}

func (l *Locator) OnIngest(event IngestEvent) {
	l.logger.Debug("received ingest event", zap.Any("event", event))

	if event.Err != nil {
		return
	}

	l.Update(event.Groups)
}

func (l *Locator) Update(gps *Groups) {
	if l.logger.Level().Enabled(zapcore.DebugLevel) {
		for g := range gps.All() {
			l.logger.Debug("group",
				zap.String("name", g.Name()),
				zap.Strings("services", g.services),
			)

			for e := range g.Endpoints() {
				l.logger.Debug("endpoint",
					zap.String("group", g.Name()),
					zap.String("originalName", e.OriginalName()),
				)
			}
		}
	}

	ringsByName := make(map[string]*consistent.Ring[*Endpoint], gps.Len())
	rings := make([]*consistent.Ring[*Endpoint], 0, gps.Len())

	for group := range gps.All() {
		ring := l.builder.Build(
			group.Len(),
			medley.Objectify(
				func(s *Endpoint) string {
					return s.OriginalName()
				},
				group.Endpoints(),
			),
		)

		ringsByName[group.Name()] = ring
		rings = append(rings, ring)
	}

	l.lock.Lock()
	l.groups = gps
	l.ringsByName = ringsByName
	l.allRings = rings
	l.lock.Unlock()
}
