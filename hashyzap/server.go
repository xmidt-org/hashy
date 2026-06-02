// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashyzap

import (
	"codeberg.org/miekg/dns"
	"go.uber.org/zap"
)

// Server creates a zap field for a named dns.Server.
func Server(fieldName, serverName string, s *dns.Server) zap.Field {
	return zap.Dict(
		fieldName,
		zap.String("name", serverName),
		zap.String("addr", s.Addr),
		zap.String("net", s.Net),
	)
}
