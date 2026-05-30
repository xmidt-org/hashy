// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
)

// IngestEvent holds information about an update to the set of groups.
type IngestEvent struct {
	// Err contains any error that occurred. If this is non-nil, Groups
	// should be ignored.
	Err error

	// Lists holds the ingested groups.
	Groups *Groups
}

type IngestListener interface {
	OnIngest(IngestEvent)
}

// Ingester represents the behavior of something that can read in group information.
// Primarily, this will be from some source of DNS RRs.
type Ingester interface {
	// AddIngestListeners adds one or more listeners. This method does not dedupe. Listeners added more
	// than once will receive duplicate events.
	AddIngestListeners(...IngestListener)

	// RemoveIngestListeners removes one or more listeners. If any listeners are not present, they are
	// ignored. If a given listener was added multiple times, it will need to be removed exactly those number
	// of times in order to stop receiving events.
	RemoveIngestListeners(...IngestListener)

	// Ingest reads group information, usually DNS RRs, from the configured source.
	// This method dispatches an IngestEvent that will contain the new groups as well
	// as any error that occurred.
	Ingest(context.Context)
}
