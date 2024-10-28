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

// Client represents the configuration for the DNS client
// hashy uses to relay DNS requests.
type Client struct {
	Network string `json:"network" yaml:"network" mapstructure:"network"`

	Timeout      time.Duration `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	DialTimeout  time.Duration `json:"dialTimeout" yaml:"dialTimeout" mapstructure:"dialTimeout"`
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`

	UDPSize uint16 `json:"udpSize" yaml:"udpSize" mapstructure:"udpSize"`
}

func (c Client) NewClient() *dns.Client {
	return &dns.Client{
		Net:          c.Network,
		Timeout:      c.Timeout,
		DialTimeout:  c.DialTimeout,
		ReadTimeout:  c.ReadTimeout,
		WriteTimeout: c.WriteTimeout,
		UDPSize:      c.UDPSize,
	}
}

// Config represents the configuration file or document that sets
// up a hashy process.  This is the top-level configuration object that
// is unmarshaled.
type Config struct {
	// Servers configures the set of dns.Server instances that get started.
	// If this field is unset, (2) dns.Server instances get started:
	// (1) a udp server on the DefaultDNSAddress, and (2) a tcp server on the DefaultDNSAddress.
	Servers Servers `json:"servers" yaml:"servers" mapstructure:"servers"`

	// Client configures the dns.Client used to relay DNS requests not
	// handled by hashy.
	Client Client `json:"client" yaml:"client" mapstructure:"client"`
}
