package service

import (
	"sync"

	"github.com/xmidt-org/medley"
	"github.com/xmidt-org/medley/consistent"
)

// Locator is a service locator backed by one or more medley consistent hash Rings.
type Locator struct {
	builder *consistent.Builder[string, *Endpoint]

	lock        sync.RWMutex
	groups      *Groups
	ringsByName map[string]*consistent.Ring[*Endpoint]
	rings       []*consistent.Ring[*Endpoint]
}

func (l *Locator) OnIngest(event IngestEvent) {
	if event.Err != nil {
		return
	}

	l.Update(event.Groups)
}

func (l *Locator) Find(object []byte) []*Endpoint {
	defer l.lock.RUnlock()
	l.lock.RLock()

	results := make([]*Endpoint, len(l.rings))
	for i, ring := range l.rings {
		results[i] = ring.Nearest(object)
	}

	return results
}

func (l *Locator) FindString(object string) []*Endpoint {
	defer l.lock.RUnlock()
	l.lock.RLock()

	results := make([]*Endpoint, len(l.rings))
	for i, ring := range l.rings {
		results[i] = ring.NearestString(object)
	}

	return results
}

func (l *Locator) Update(gps *Groups) {
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
