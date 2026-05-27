package server

import (
	"fmt"

	"codeberg.org/miekg/dns"
	"github.com/xmidt-org/hashy/config"
	"go.uber.org/zap"
)

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

// NewDNSServers creates all the DNS servers listed in the configuration, creates handlers for each one,
// and binds each server to the enclosing fx.App lifecycle.
func NewDNSServers(middleware *Middleware, lifecycler *Lifecycler, cfg config.DNS) (dnsServers []*dns.Server, err error) {
	dnsServers = make([]*dns.Server, 0, len(cfg.UDP)+len(cfg.TCP))
	for name, udpConfig := range cfg.UDP {
		var server *dns.Server
		server, err = NewUDPServer(udpConfig)
		if err != nil {
			return
		}

		var serverLogger *zap.Logger
		server.Handler, serverLogger = middleware.Then(name)
		lifecycler.Append(serverLogger, server)
		dnsServers = append(dnsServers, server)
	}

	for name, tcpConfig := range cfg.TCP {
		var server *dns.Server
		server, err = NewTCPServer(tcpConfig)
		if err != nil {
			return
		}

		var serverLogger *zap.Logger
		server.Handler, serverLogger = middleware.Then(name)
		lifecycler.Append(serverLogger, server)
		dnsServers = append(dnsServers, server)
	}

	return
}
