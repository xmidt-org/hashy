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

func (le LocatedEndpoints) emptyRRs(func(*Endpoint, dns.RR) bool) {}

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
		return le.emptyRRs
	}
}

// Locator is a service locator backed by one or more medley consistent hash Rings.
type Locator struct {
	logger  *zap.Logger
	builder *consistent.Builder[string, *Endpoint]

	lock        sync.RWMutex
	groups      *Groups
	ringsByName map[string]*consistent.Ring[*Endpoint]
	rings       []*consistent.Ring[*Endpoint]
}

func (l *Locator) Find(object []byte) LocatedEndpoints {
	defer l.lock.RUnlock()
	l.lock.RLock()

	results := make(LocatedEndpoints, len(l.rings))
	for i, ring := range l.rings {
		results[i] = ring.Nearest(object)
	}

	return results
}

func (l *Locator) FindString(object string) LocatedEndpoints {
	defer l.lock.RUnlock()
	l.lock.RLock()

	results := make(LocatedEndpoints, len(l.rings))
	for i, ring := range l.rings {
		results[i] = ring.NearestString(object)
	}

	return results
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
			for e := range g.Endpoints() {
				l.logger.Debug("endpoint",
					zap.String("group", g.Name()),
					zap.Dict("endpoint",
						zap.String("name", e.Name()),
						zap.String("originalName", e.OriginalName()),
					),
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
					return s.Name()
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
	l.rings = rings
	l.lock.Unlock()
}
