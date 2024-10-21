package hashy

import "time"

// ServerConfig represents a single server's configuration within
// the hashy process.
type ServerConfig struct {
	Name    string `json:"name" yaml:"name" mapstructure:"name"`
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

// ClientConfig represents the configuration for the DNS client
// hashy uses to relay DNS requests.
type ClientConfig struct {
	Network string `json:"network" yaml:"network" mapstructure:"network"`

	Timeout      time.Duration `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	DialTimeout  time.Duration `json:"dialTimeout" yaml:"dialTimeout" mapstructure:"dialTimeout"`
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`

	UDPSize int `json:"udpSize" yaml:"udpSize" mapstructure:"udpSize"`
}

// HashConfig holds the configuration for the hashing algorithm used when
// hashing matching DNS names.
type HashConfig struct {
	Vnodes int `json:"vnodes" yaml:"vnodes" mapstructure:"vnodes"`
}

// Config represents the configuration file or document that sets
// up a hashy server.
type Config struct {
	Servers []ServerConfig `json:"servers" yaml:"servers" mapstructure:"servers"`
	Client  ClientConfig   `json:"client" yaml:"client" mapstructure:"client"`
	Hash    HashConfig     `json:"hash" yaml:"hash" mapstructure:"hash"`
}
