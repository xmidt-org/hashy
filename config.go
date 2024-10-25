package hashy

import (
	"time"

	"github.com/miekg/dns"
)

// ServerConfig represents a single server's configuration within
// the hashy process.
type ServerConfig struct {
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

func (sc ServerConfig) NewServer() *dns.Server {
	return &dns.Server{
		Addr:          sc.Address,
		Net:           serverNetwork(sc.Network),
		ReadTimeout:   sc.ReadTimeout,
		WriteTimeout:  sc.WriteTimeout,
		IdleTimeout:   serverIdleTimeout(sc.IdleTimeout),
		UDPSize:       sc.UDPSize,
		MaxTCPQueries: sc.MaxTCPQueries,
		ReusePort:     sc.ReusePort,
		ReuseAddr:     sc.ReuseAddress,
	}
}

// ServerConfigs is an aggregate of multiple server configurations.
type ServerConfigs []ServerConfig

func (scs ServerConfigs) NewServers() (s []*dns.Server) {
	s = make([]*dns.Server, 0, len(scs))
	for _, sc := range scs {
		s = append(s, sc.NewServer())
	}

	return
}

// CacheConfig holds the configuration for the client cache of DNS responses.
type CacheConfig struct {
	MaxEntries int `json:"maxEntries" yaml:"maxEntries" mapstructure:"maxEntries"`
}

// ClientConfig represents the configuration for the DNS client
// hashy uses to relay DNS requests.
type ClientConfig struct {
	Network string `json:"network" yaml:"network" mapstructure:"network"`

	Timeout      time.Duration `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	DialTimeout  time.Duration `json:"dialTimeout" yaml:"dialTimeout" mapstructure:"dialTimeout"`
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`

	UDPSize uint16 `json:"udpSize" yaml:"udpSize" mapstructure:"udpSize"`

	Cache CacheConfig `json:"cache" yaml:"cache" mapstructure:"cache"`
}

func (cc ClientConfig) NewClient() *dns.Client {
	return &dns.Client{
		Net:          cc.Network,
		Timeout:      cc.Timeout,
		DialTimeout:  cc.DialTimeout,
		ReadTimeout:  cc.ReadTimeout,
		WriteTimeout: cc.WriteTimeout,
		UDPSize:      cc.UDPSize,
	}
}

// HashConfig holds the configuration for the hashing algorithm used when
// hashing matching DNS names.
type HashConfig struct {
	DNSSuffixes []string `json:"dnsSuffixes" yaml:"dnsSuffixes" mapstructure:"dnsSuffixes"`
	Vnodes      int      `json:"vnodes" yaml:"vnodes" mapstructure:"vnodes"`
}

// Config represents the configuration file or document that sets
// up a hashy server.
type Config struct {
	Servers ServerConfigs `json:"servers" yaml:"servers" mapstructure:"servers"`
	Client  ClientConfig  `json:"client" yaml:"client" mapstructure:"client"`
	Hash    HashConfig    `json:"hash" yaml:"hash" mapstructure:"hash"`
}
