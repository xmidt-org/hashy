// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	"codeberg.org/miekg/dns"
	"github.com/xmidt-org/hashy/config"
	"github.com/xmidt-org/hashy/hashyzap"
	"go.uber.org/zap"
)

// NewServerLogger produces a sublogger appropriate for server-specific messages.
func NewServerLogger(parent *zap.Logger, serverName string, server *dns.Server) *zap.Logger {
	return parent.With(
		hashyzap.Server("server", serverName, server),
	)
}

// NewUDPServer creates a *dns.Server from configuration.
func NewUDPServer(cfg config.UDP) (s *dns.Server, err error) {
	s = dns.NewServer()

	if len(cfg.Address) > 0 {
		s.Addr = cfg.Address
	}

	switch cfg.Network {
	case "udp", "udp4", "udp6":
		s.Net = cfg.Network

	case "":
		s.Net = "udp"

	default:
		return nil, fmt.Errorf("network for a udp server must be either blank or one of: [udp, udp4, udp6]")
	}

	if cfg.Size > 0 {
		s.UDPSize = cfg.Size
	}

	if cfg.ReadTimeout > 0 {
		s.ReadTimeout = cfg.ReadTimeout
	}

	if cfg.IdleTimeout > 0 {
		s.IdleTimeout = cfg.IdleTimeout
	}

	s.ReusePort = cfg.ReusePort
	s.ReuseAddr = cfg.ReusePort

	return
}

// NewTCPServer creates a *dns.Server from configuration.
func NewTCPServer(cfg config.TCP) (s *dns.Server, err error) {
	s = dns.NewServer()

	if len(cfg.Address) > 0 {
		s.Addr = cfg.Address
	}

	switch cfg.Network {
	case "tcp", "tcp4", "tcp6":
		s.Net = cfg.Network

	case "":
		s.Net = "tcp"

	default:
		return nil, fmt.Errorf("network for a tcp server must be either blank or one of: [tcp, tcp4, tcp6]")
	}

	if cfg.MaxQueries > 0 {
		s.MaxTCPQueries = cfg.MaxQueries
	}

	if cfg.ReadTimeout > 0 {
		s.ReadTimeout = cfg.ReadTimeout
	}

	if cfg.IdleTimeout > 0 {
		s.IdleTimeout = cfg.IdleTimeout
	}

	s.ReusePort = cfg.ReusePort
	s.ReuseAddr = cfg.ReusePort

	return
}
