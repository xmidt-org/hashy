// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/config"
	"go.uber.org/zap"
)

const (
	// DefaultFileIngesterOrigin is the default $ORIGIN used for parsing zone files.
	DefaultFileIngesterOrigin = ""

	// DefaultFileIngesterTTL is the default $TTL used for parsing zone files.
	DefaultFileIngesterTTL = 5 * time.Minute
)

type FileIngesterOption interface {
	applyToFileIngester(*FileIngester) error
}

type fileIngesterOptionFunc func(*FileIngester) error

func (f fileIngesterOptionFunc) applyToFileIngester(fi *FileIngester) error { return f(fi) }

func WithIngestLogger(base *zap.Logger) FileIngesterOption {
	return fileIngesterOptionFunc(func(fi *FileIngester) error {
		if base == nil {
			base = zap.NewNop()
		}

		fi.logger = base.Named("fileIngester")
		return nil
	})
}

func WithGlobs(more ...string) FileIngesterOption {
	return fileIngesterOptionFunc(func(fi *FileIngester) error {
		fi.globs = slices.Grow(fi.globs, len(more))
		for _, m := range more {
			fi.globs = append(fi.globs, os.ExpandEnv(m))
		}

		return nil
	})
}

func WithOrigin(origin string) FileIngesterOption {
	return fileIngesterOptionFunc(func(fi *FileIngester) error {
		fi.origin = origin
		return nil
	})
}

func WithTTL(ttl time.Duration) FileIngesterOption {
	return fileIngesterOptionFunc(func(fi *FileIngester) error {
		fi.ttl = hashy.DurationToSeconds(ttl)
		return nil
	})
}

func WithDiscoveryDomain(domain string) FileIngesterOption {
	return fileIngesterOptionFunc(func(fi *FileIngester) error {
		if len(domain) > 0 {
			fi.discoveryDomain = dnsutil.Fqdn(domain)
		} else {
			fi.discoveryDomain = ""
		}

		return nil
	})
}

func WithIngestListeners(more ...IngestListener) FileIngesterOption {
	return fileIngesterOptionFunc(func(fi *FileIngester) error {
		fi.listeners = slices.Grow(fi.listeners, len(more))
		fi.listeners = append(fi.listeners, more...)
		return nil
	})
}

func WithGroupsConfig(gcfg config.Groups) FileIngesterOption {
	return fileIngesterOptionFunc(func(fi *FileIngester) (err error) {
		err = WithDiscoveryDomain(gcfg.DiscoveryDomain).
			applyToFileIngester(fi)

		if err == nil {
			err = WithGlobs(gcfg.ZoneFiles...).
				applyToFileIngester(fi)
		}

		if err == nil {
			err = WithOrigin(gcfg.Origin).
				applyToFileIngester(fi)
		}

		if err == nil {
			err = WithTTL(gcfg.DefaultTTL).
				applyToFileIngester(fi)
		}

		return
	})
}

// FileIngester handles reading in DNS zone files from the filesystem.
type FileIngester struct {
	logger          *zap.Logger
	globs           []string
	origin          string
	ttl             uint32
	discoveryDomain string

	lock      sync.RWMutex
	listeners []IngestListener
}

// NewFileIngester creates a FileIngester from a set of options.
func NewFileIngester(opts ...FileIngesterOption) (*FileIngester, error) {
	fi := new(FileIngester)
	for _, o := range opts {
		if err := o.applyToFileIngester(fi); err != nil {
			return nil, err
		}
	}

	if fi.logger == nil {
		fi.logger = zap.NewNop()
	}

	if len(fi.origin) == 0 {
		fi.origin = DefaultFileIngesterOrigin
	}

	if fi.ttl == 0 {
		fi.ttl = hashy.DurationToSeconds(DefaultFileIngesterTTL)
	}

	if len(fi.discoveryDomain) == 0 {
		// we know this won't cause an error
		WithDiscoveryDomain(DefaultDiscoveryDomain).
			applyToFileIngester(fi)
	}

	return fi, nil
}

func (fi *FileIngester) AddIngestListeners(more ...IngestListener) {
	fi.lock.Lock()
	fi.listeners = slices.Grow(fi.listeners, len(more))
	fi.listeners = append(fi.listeners, more...)
	fi.lock.Unlock()
}

func (fi *FileIngester) RemoveIngestListeners(less ...IngestListener) {
	fi.lock.Lock()
	defer fi.lock.Unlock()

	fi.listeners = slices.DeleteFunc(fi.listeners, func(candidate IngestListener) bool {
		return slices.Contains(less, candidate)
	})
}

// zoneParsers is a sequence of file paths for each configured glob. If any error occurs,
// the yield function is called with a an empty string and the error.
//
// Each glob's set of matching files is sorted lexicographically for a consistent
// processing order.
func (fi *FileIngester) zoneFiles(yield func(string, error) bool) {
	for _, glob := range fi.globs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			yield("", err)
			return
		}

		sort.Strings(matches)
		for _, path := range matches {
			if !yield(path, nil) {
				return
			}
		}
	}
}

// newZoneParser creates a *ZoneParser for a path. If the context has expired, or if a problem
// occurs opening the path, and error is returned and the other returned values will be nil.
func (fi *FileIngester) newZoneParser(ctx context.Context, path string) (parser *dns.ZoneParser, closer io.Closer, err error) {
	if ctx.Err() != nil {
		err = ctx.Err()
		return
	}

	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return
	}

	closer = file
	parser = dns.NewZoneParser(file, fi.origin, path)
	parser.IncludeFS = os.DirFS(filepath.Dir(path)) // includes are relative to the location of the file
	parser.SetDefaultTTL(fi.ttl)
	return
}

func (fi *FileIngester) ingestFile(ctx context.Context, l *zap.Logger, rrc *RRCollector, path string) error {
	parser, closer, err := fi.newZoneParser(ctx, path)
	if err != nil {
		return err
	}

	defer closer.Close()
	for rr, err := range parser.RRs() {
		if rr == nil {
			// a successful end is a nil RR and a nil error
			// otherwise, err will hold any error that occurred
			return err
		}

		l.Debug("resource record", zap.Stringer("rr", rr))
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

func (fi *FileIngester) Ingest(ctx context.Context) {
	var event IngestEvent
	rrc := RRCollector{
		discoveryDomain: fi.discoveryDomain,
	}

	fileCount := 0
	for path, err := range fi.zoneFiles {
		if err != nil {
			event.Err = err
			fi.logger.Error("failed to expand file globs", zap.Error(event.Err))
			break
		}

		fileCount++
		ingestLogger := fi.logger.With(zap.String("path", path))
		ingestLogger.Debug("parsing zone file")

		event.Err = fi.ingestFile(
			ctx,
			ingestLogger,
			&rrc,
			path,
		)

		if event.Err != nil {
			fi.logger.Error("error parsing file", zap.String("path", path), zap.Error(err))
			break
		}
	}

	fi.logger.Info("parsing complete", zap.Int("fileCount", fileCount))
	if event.Err == nil {
		event.Groups = rrc.Build()
	}

	fi.dispatchIngestEvent(event)
}
