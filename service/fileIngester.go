// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"hash/adler32"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync/atomic"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/config"
	"github.com/xmidt-org/medley"
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

	checksummer medley.Constructor[uint32]
	first       atomic.Bool
	checksum    atomic.Uint32

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

	if fi.checksummer == nil {
		fi.checksummer = medley.AsConstructor32(adler32.New)
	}

	return fi, nil
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
func (fi *FileIngester) newZoneParser(checksummer medley.Hash[uint32], path string) (parser *dns.ZoneParser, closer io.Closer, err error) {
	var file *os.File
	file, err = os.Open(path)
	if err != nil {
		return
	}

	closer = file
	reader := io.TeeReader(file, checksummer)
	parser = dns.NewZoneParser(reader, fi.origin, path)
	parser.IncludeFS = os.DirFS(filepath.Dir(path)) // includes are relative to the location of the file
	parser.SetDefaultTTL(fi.ttl)
	return
}

func (fi *FileIngester) ingestFile(l *zap.Logger, checksummer medley.Hash[uint32], rrc *RRCollector, path string) error {
	parser, closer, err := fi.newZoneParser(checksummer, path)
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
	for _, l := range fi.listeners {
		l.OnIngest(event)
	}
}

// Ingest reads in all the files this FileIngester was configured with.
// This method tracks the checksum across all files. An IngestEvent will
// only be dispatch if either (a) this is the first Ingest, or (b) if any
// change in the files occurred.
func (fi *FileIngester) Ingest(ctx context.Context) {
	var event IngestEvent
	rrc := RRCollector{
		discoveryDomain: fi.discoveryDomain,
	}

	oldChecksum := fi.checksum.Load()
	checksummer := fi.checksummer()
	fileCount := 0

	for path, err := range fi.zoneFiles {
		if err != nil {
			event.Err = err
			fi.logger.Error("failed to expand file globs", zap.Error(event.Err))
			break
		}

		event.Err = ctx.Err()
		if event.Err != nil {
			break
		}

		fileCount++
		ingestLogger := fi.logger.With(zap.String("path", path))
		ingestLogger.Debug("parsing zone file")
		event.Err = fi.ingestFile(ingestLogger, checksummer, &rrc, path)

		if event.Err != nil {
			fi.logger.Error("error parsing file", zap.String("path", path), zap.Error(err))
			break
		}
	}

	fi.logger.Info("parsing complete", zap.Int("fileCount", fileCount))
	if event.Err == nil {
		newChecksum := checksummer.Value()

		if fi.first.CompareAndSwap(false, true) {
			// on the first time we ingest, we always build groups and dispatch
			fi.logger.Info("initial ingest")
			event.Groups = rrc.Build()
			fi.checksum.Store(newChecksum)
		} else if fi.checksum.Load() != newChecksum && fi.checksum.CompareAndSwap(oldChecksum, newChecksum) {
			// only build groups if there was a change that we recognize
			fi.logger.Info("changes detected")
			event.Groups = rrc.Build()
		}

		if event.Groups != nil {
			fi.dispatchIngestEvent(event)
		} else {
			fi.logger.Info("no changes since last ingest")
		}
	} else {
		// always dispatch errors
		fi.dispatchIngestEvent(event)
	}
}
