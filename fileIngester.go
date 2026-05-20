// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashy

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"

	"codeberg.org/miekg/dns"
)

type FileIngester struct {
	globs           []string
	origin          string
	defaultTTL      uint32
	discoveryDomain string
	nameGenerator   ServerNameGenerator

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

func (fi *FileIngester) ingestFile(ctx context.Context, rrc *RRCollector, path string) error {
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

		if err := rrc.Add(rr); err != nil {
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
	rrc := RRCollector{
		discoveryDomain: fi.discoveryDomain,
		nameGenerator:   fi.nameGenerator,
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
		event.Groups = rrc.Build()
	}

	fi.dispatchIngestEvent(event)
}
