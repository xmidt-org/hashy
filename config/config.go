// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package config

import (
	_ "embed"
	"time"

	"github.com/xmidt-org/sallust"
)

// Default is the server's default configuration.
//
//go:embed defaultConfig.yaml
var Default string

// Zone describes the zone that hashy serves.
type Zone struct {
	// Domain is the domain that hash serves. If unset, this defaults to DefaultDomain.
	Domain string `json:"domain" yaml:"domain" mapstructure:"domain"`

	// TTL is the base time-to-live of records generated in this zone.
	TTL time.Duration `json:"ttl" yaml:"ttl" mapstructure:"ttl"`

	// TTLJitter is the percentage of variance for the TTL. Each DNS request has the TTL
	// varied by this value. The default is 0, meaning no jitter.
	//
	// If this value is outside the range 0.0 <= TTLJitter < 1.0, an error is raised.
	TTLJitter float32 `json:"ttlJitter" yaml:"ttlJitter" mapstructure:"ttlJitter"`
}

// UDP is the configuration for a single UDP server that serve DNS traffic.
type UDP struct {
	Address     string        `json:"address" yaml:"address" mapstructure:"address"`
	Network     string        `json:"network" yaml:"network" mapstructure:"network"`
	Size        int           `json:"size" yaml:"size" mapstructure:"size"`
	ReadTimeout time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	IdleTimeout time.Duration `json:"idleTimeout" yaml:"idleTimeout" mapstructure:"idleTimeout"`
	ReusePort   bool          `json:"reusePort" yaml:"reusePort" mapstructure:"reusePort"`
	ReuseAddr   bool          `json:"reuseAddr" yaml:"reuseAddr" mapstructure:"reuseAddr"`
}

type UDPServers map[string]UDP

// TCP is the configuration for a single TCP server that serve DNS traffic.
type TCP struct {
	Address     string        `json:"address" yaml:"address" mapstructure:"address"`
	Network     string        `json:"network" yaml:"network" mapstructure:"network"`
	MaxQueries  int           `json:"maxQueries" yaml:"maxQueries" mapstructure:"maxQueries"`
	ReadTimeout time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	IdleTimeout time.Duration `json:"idleTimeout" yaml:"idleTimeout" mapstructure:"idleTimeout"`
	ReusePort   bool          `json:"reusePort" yaml:"reusePort" mapstructure:"reusePort"`
	ReuseAddr   bool          `json:"reuseAddr" yaml:"reuseAddr" mapstructure:"reuseAddr"`
}

type TCPServers map[string]TCP

// DNS is the configuration all all servers that serve DNS traffic.
type DNS struct {
	// Zone holds information about the synthetic zone that hashy serves.
	Zone Zone `json:"zone" yaml:"zone" mapstructure:"zone"`

	// UDP holds all the UDP servers for DNS. The keys in the map are human-friendly server names.
	UDP UDPServers `json:"udp" yaml:"udp" mapstructure:"udp"`

	// TCP holds all the TCP servers for DNS. The keys in the map are human-friendly server names.
	TCP TCPServers `json:"tcp" yaml:"tcp" mapstructure:"tcp"`
}

// Groups holds the configuration necessary to establish hashy's groups.
type Groups struct {
	// DiscoveryDomain is the domain hashy queries to discover group information. If unset, this defaults to
	// DefaultDiscoveryDomain.
	DiscoveryDomain string `json:"discoveryDomain" yaml:"discoveryDomain" mapstructure:"discoveryDomain"`

	// VNodes is the number of virtual nodes to use in consistent hashing.
	VNodes int `json:"vnodes" yaml:"vnodes" mapstructure:"vnodes"`

	// CheckInterval is the interval on which external sources of DNS RRs are rechecked
	// to see if hashy may update its state. If unset, a service.DefaultCheckInterval is used.
	CheckInterval time.Duration `json:"checkInterval" yaml:"checkInterval" mapstructure:"checkInterval"`

	// ZoneFiles is a list of filesystem globs that contain group information.
	ZoneFiles []string `json:"zoneFiles" yaml:"zoneFiles" mapstructure:"zoneFiles"`

	// Origin is the origin to use when parsing zone files.
	Origin string `json:"origin" yaml:"origin" mapstructure:"origin"`

	// DefaultTTL is the default TTL to use when parsing zone files.
	DefaultTTL time.Duration `json:"defaultTTL" yaml:"defaultTTL" mapstructure:"defaultTTL"`
}

// Main is the top-level configuration object for hashy.
type Main struct {
	// DNS holds all the information about the zone and the servers.
	DNS DNS `json:"dns" yaml:"dns" mapstructure:"dns"`

	// Groups defines how hashy obtains its groups.
	Groups Groups `json:"groups" yaml:"groups" mapstructure:"groups"`

	// Logging is the server logging configuration.
	Logging sallust.Config `json:"logging" yaml:"logging" mapstructure:"logging"`
}
