// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashycfg

import (
	"time"

	"github.com/miekg/dns"
)

const (
	// DefaultServerAddress is the address used by a Server when the Address
	// field is unset.
	DefaultServerAddress = ":53"

	// DefaultServerAddress is the network used by a Server when the
	// Network field is unset.
	DefaultServerNetwork = "udp"
)

// Server represents a single server's configuration within
// the hashy process.
type Server struct {
	Address string `json:"address" yaml:"address" mapstructure:"address"`
	Network string `json:"network" yaml:"network" mapstructure:"network"`

	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `json:"idleTimeout" yaml:"idleTimeout" mapstructure:"idleTimeout"`

	UDPSize       int  `json:"udpSize" yaml:"udpSize" mapstructure:"udpSize"`
	MaxTCPQueries int  `json:"maxTCPQueries" yaml:"maxTCPQueries" mapstructure:"maxTCPQueries"`
	ReusePort     bool `json:"reusePort" yaml:"reusePort" mapstructure:"reusePort"`
	ReuseAddress  bool `json:"reuseAddress" yaml:"reuseAddress" mapstructure:"reuseAddress"`
}

func (s Server) address() string {
	if len(s.Address) > 0 {
		return s.Address
	}

	return DefaultServerAddress
}

func (s Server) network() string {
	if len(s.Network) > 0 {
		return s.Network
	}

	return DefaultServerNetwork
}

// NewServer creates a dns.Server from this configuration.
func (s Server) NewServer() *dns.Server {
	server := &dns.Server{
		Addr:          s.address(),
		Net:           s.network(),
		ReadTimeout:   s.ReadTimeout,
		WriteTimeout:  s.WriteTimeout,
		UDPSize:       s.UDPSize,
		MaxTCPQueries: s.MaxTCPQueries,
		ReusePort:     s.ReusePort,
		ReuseAddr:     s.ReuseAddress,
	}

	if s.IdleTimeout > 0 {
		server.IdleTimeout = func() time.Duration {
			return s.IdleTimeout
		}
	}

	return server
}

// Servers is an aggregate of multiple server configurations.
type Servers []Server

// NewServers creates a slice of dns.Server instances corresponding to this
// configuration.
func (ss Servers) NewServers() (servers []*dns.Server) {
	if len(ss) > 0 {
		servers = make([]*dns.Server, 0, len(ss))
		for _, s := range ss {
			servers = append(servers, s.NewServer())
		}
	} else {
		// by default, start a udp and a tcp server on the default UDP port
		servers = []*dns.Server{
			Server{}.NewServer(),
			Server{Network: "tcp"}.NewServer(),
		}
	}

	return
}

// ZoneConfig holds any static RR records that hashy should know about.
type ZoneConfig struct {
	// Origin is the origin name for hashy.  This is used as the $ORIGIN
	// when parsing master files.
	Origin string `json:"origin" yaml:"origin" mapstructure:"origin"`

	// Files is a list of system paths that are RFC1035 master files.
	// These files can contain an SOA record for hashy's domain as well
	// as extra domain name information that hashy will use when responding
	// to DNS requests.
	Files []string `json:"files" yaml:"files" mapstructure:"files"`

	// Text is an embedded set of RR records.  This field allows a master
	// file to be embedded within hashy's configuration file.
	Text string `json:"text" yaml:"text" mapstructure:"text"`
}

// Config represents the configuration file or document that sets
// up a hashy process.  This is the top-level configuration object that
// is unmarshaled.
type Config struct {
	// Servers configures the set of dns.Server instances that get started.
	// If this field is unset, (2) dns.Server instances get started:
	// (1) a udp server on the DefaultServerAddress, and (2) a tcp server
	// on the DefaultServerAddress.
	Servers Servers `json:"servers" yaml:"servers" mapstructure:"servers"`

	// Zone holds information about the DNS zone that hashy serves.
	Zone ZoneConfig `json:"zone" yaml:"zone" mapstructure:"zone"`
}
