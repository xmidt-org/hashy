// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	// DefaultCheckInterval is the default time that a FileIngester rechecks external sources
	// for DNS RRs that affect how hashy operates.
	DefaultCheckInterval = 10 * time.Minute
)

// IngestEvent holds information about an update to the set of groups.
type IngestEvent struct {
	// Err contains any error that occurred. If this is non-nil, Groups
	// should be ignored.
	Err error

	// Lists holds the ingested groups.
	Groups *Groups
}

// IngestListener is a sink for IngestEvents.
type IngestListener interface {
	// OnIngest notifies this listener that an ingest operation has completed.
	// Only errors and actual changes will be dispatched through this method.
	OnIngest(IngestEvent)
}

// Ingester represents the behavior of something that can read in group information.
// Primarily, this will be from some source of DNS RRs.
type Ingester interface {
	// Ingest reads group information, usually DNS RRs, from the configured source.
	// This method dispatches an IngestEvent that will contain the new groups as well
	// as any error that occurred.
	Ingest(context.Context)
}

type IngestCheckerOption interface {
	applyToIngestChecker(*IngestChecker) error
}

type ingestCheckerOptionFunc func(*IngestChecker) error

func (f ingestCheckerOptionFunc) applyToIngestChecker(ic *IngestChecker) error { return f(ic) }

func WithCheckInterval(v time.Duration) IngestCheckerOption {
	return ingestCheckerOptionFunc(func(ic *IngestChecker) error {
		ic.interval = v
		return nil
	})
}

func WithIngester(i Ingester) IngestCheckerOption {
	return ingestCheckerOptionFunc(func(ic *IngestChecker) error {
		ic.ingester = i
		return nil
	})
}

// IngestChecker manages a single goroutine that invokes Ingest on a particular Ingester
// on an interval.
//
// Only (1) background goroutine will run for any given IngestChecker.
type IngestChecker struct {
	interval time.Duration
	ingester Ingester

	runLock    sync.Mutex
	cancelFunc context.CancelFunc
}

// NewIngestChecker creates an unstarted IngestChecker using the supplied options.
// If no Ingester was supplied in the options, this method returns an error.
func NewIngestChecker(opts ...IngestCheckerOption) (*IngestChecker, error) {
	ic := new(IngestChecker)
	for _, o := range opts {
		if err := o.applyToIngestChecker(ic); err != nil {
			return nil, err
		}
	}

	if ic.interval <= 0 {
		ic.interval = DefaultCheckInterval
	}

	if ic.ingester == nil {
		return nil, errors.New("an Ingester is required for an IngestChecker")
	}

	return ic, nil
}

// run is a goroutine that simply invokes Ingest on an interval until
// the context is canceled.
//
// Since an IngestChecker is immutable, this method does not need to
// contend on a lock to use state.
func (ic *IngestChecker) run(ctx context.Context) {
	ticker := time.NewTicker(ic.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			ic.ingester.Ingest(ctx)
		}
	}
}

// Start atomically starts invoking Ingest on the configured interval.
// This method is idempotent.
func (ic *IngestChecker) Start() {
	defer ic.runLock.Unlock()
	ic.runLock.Lock()

	if ic.cancelFunc == nil {
		var ctx context.Context
		ctx, ic.cancelFunc = context.WithCancel(context.Background())
		go ic.run(ctx)
	}
}

// Stop atomically halts invoking Ingest. This method is idempotent.
func (ic *IngestChecker) Stop() {
	defer ic.runLock.Unlock()
	ic.runLock.Lock()

	if ic.cancelFunc != nil {
		ic.cancelFunc()
		ic.cancelFunc = nil
	}
}
