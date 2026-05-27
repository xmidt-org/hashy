// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"iter"
)

// Group is a single group of servers.
type Group struct {
	name      string
	endpoints []Endpoint
}

func (g *Group) Len() int {
	return len(g.endpoints)
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Endpoints() iter.Seq[*Endpoint] {
	return func(yield func(*Endpoint) bool) {
		for i := range len(g.endpoints) {
			if !yield(&g.endpoints[i]) {
				return
			}
		}
	}
}

// Groups is an immutable collection of List instances.
type Groups struct {
	byName map[string]int
	all    []Group
}

func (gps *Groups) Len() int {
	return len(gps.all)
}

func (gps *Groups) At(i int) *Group {
	return &gps.all[i]
}

func (gps *Groups) Get(groupName string) *Group {
	if pos, existing := gps.byName[groupName]; existing {
		return &gps.all[pos]
	}

	return nil
}

func (gps *Groups) All() iter.Seq[*Group] {
	return func(yield func(*Group) bool) {
		for i := range len(gps.all) {
			if !yield(&gps.all[i]) {
				return
			}
		}
	}
}
