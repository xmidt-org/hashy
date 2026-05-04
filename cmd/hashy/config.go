// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	_ "embed"
	"time"
)

//go:embed defaultConfig.yaml
var defaultConfig string

// ConfigLocation is a marker type for a component which holds where hashy's
// configuration was loaded from.
type ConfigLocation string

// UDPServerConfig is the configuration for a single UDP server that serve DNS traffic.
type UDPServerConfig struct {
	Address      string        `json:"address" yaml:"address" mapstructure:"address"`
	Size         int           `json:"size" yaml:"size" mapstructure:"size"`
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `json:"idleTimeout" yaml:"idleTimeout" mapstructure:"idleTimeout"`
	ReusePort    bool          `json:"reusePort" yaml:"reusePort" mapstructure:"reusePort"`
	ReuseAddr    bool          `json:"reuseAddr" yaml:"reuseAddr" mapstructure:"reuseAddr"`
}

// TCPServerConfig is the configuration for a single TCP server that serve DNS traffic.
type TCPServerConfig struct {
	Address      string        `json:"address" yaml:"address" mapstructure:"address"`
	MaxQueries   int           `json:"maxQueries" yaml:"maxQueries" mapstructure:"maxQueries"`
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `json:"idleTimeout" yaml:"idleTimeout" mapstructure:"idleTimeout"`
	ReusePort    bool          `json:"reusePort" yaml:"reusePort" mapstructure:"reusePort"`
	ReuseAddr    bool          `json:"reuseAddr" yaml:"reuseAddr" mapstructure:"reuseAddr"`
}

// DNSServersConfig is the configuration all all servers that serve DNS traffic.
type DNSServersConfig struct {
	// UDP holds all the UDP servers for DNS. The keys in the map are human-friendly server names.
	UDP map[string]UDPServerConfig `json:"udp" yaml:"udp" mapstructure:"udp"`

	// TCP holds all the TCP servers for DNS. The keys in the map are human-friendly server names.
	TCP map[string]TCPServerConfig `json:"tcp" yaml:"tcp" mapstructure:"tcp"`
}

// ServersConfig holds the configuration for all servers that hashy starts, both DNS and any HTTP servers.
type ServersConfig struct {
	DNS DNSServersConfig `json:"dns" yaml:"dns" mapstructure:"dns"`
}

// Config is the top-level configuration object for the hashy server.
type Config struct {
	// Servers holds all the hashy server configurations.
	Servers ServersConfig `json:"servers" yaml:"servers" mapstructure:"servers"`
}
