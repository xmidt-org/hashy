// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"time"

	"codeberg.org/miekg/dns"
	"go.uber.org/zap"
)

const (
	DefaultFileIngesterOrigin = ""
	DefaultFileIngesterTTL    = 5 * time.Minute
)

type FileIngester struct {
	Logger          *zap.Logger
	ZoneFiles       []string
	Origin          string
	DefaultTTL      uint32
	DiscoveryDomain string
	NameGenerator   EndpointNameGenerator

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
	zp := dns.NewZoneParser(file, fi.Origin, path)
	zp.IncludeFS = os.DirFS(filepath.Dir(path)) // includes are relative to the location of the file
	zp.SetDefaultTTL(fi.DefaultTTL)

	for rr, err := range zp.RRs() {
		if err != nil {
			return err
		}

		if err := rrc.AddRR(rr); err != nil {
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

func (fi *FileIngester) newRRCollector() (rrc RRCollector) {
	rrc = RRCollector{
		discoveryDomain: fi.DiscoveryDomain,
		nameGenerator:   fi.NameGenerator,
	}

	if len(rrc.discoveryDomain) == 0 {
		rrc.discoveryDomain = DefaultDiscoveryDomain
	}

	return
}

func (fi *FileIngester) Ingest(ctx context.Context) {
	var event IngestEvent
	rrc := fi.newRRCollector()

	var paths []string
	for _, glob := range fi.ZoneFiles {
		glob = os.ExpandEnv(glob)
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
		fi.Logger.Info("zone files to parse", zap.Int("fileCount", len(paths)), zap.Strings("files", paths))
	}

	var pathIndex int
	for pathIndex = 0; pathIndex < len(paths) && event.Err == nil; pathIndex++ {
		path := paths[pathIndex]
		ingestLogger := fi.Logger.With(zap.String("path", path))
		ingestLogger.Debug("parsing zone file")

		event.Err = fi.ingestFile(
			ctx,
			&rrc,
			path,
		)
	}

	if event.Err == nil {
		fi.Logger.Info("parsing complete", zap.Int("fileCount", pathIndex))
		event.Groups = rrc.Build()
	} else {
		fi.Logger.Error("failed to parse zone files", zap.Error(event.Err))
	}

	fi.dispatchIngestEvent(event)
}
