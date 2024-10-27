package hashycfg

import (
	"time"

	"github.com/miekg/dns"
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

// serverNetwork returns v if it is not empty, "dns" otherwise.
func serverNetwork(v string) string {
	if len(v) > 0 {
		return v
	}

	return "udp"
}

func serverIdleTimeout(v time.Duration) (f func() time.Duration) {
	if v > 0 {
		f = func() time.Duration {
			return v
		}
	}

	return
}

func (s Server) NewServer() *dns.Server {
	return &dns.Server{
		Addr:          s.Address,
		Net:           serverNetwork(s.Network),
		ReadTimeout:   s.ReadTimeout,
		WriteTimeout:  s.WriteTimeout,
		IdleTimeout:   serverIdleTimeout(s.IdleTimeout),
		UDPSize:       s.UDPSize,
		MaxTCPQueries: s.MaxTCPQueries,
		ReusePort:     s.ReusePort,
		ReuseAddr:     s.ReuseAddress,
	}
}

// Servers is an aggregate of multiple server configurations.
type Servers []Server

func (ss Servers) NewServers() (servers []*dns.Server) {
	servers = make([]*dns.Server, 0, len(ss))
	for _, s := range ss {
		servers = append(servers, s.NewServer())
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
	Servers Servers `json:"servers" yaml:"servers" mapstructure:"servers"`
	Client  Client  `json:"client" yaml:"client" mapstructure:"client"`
}
