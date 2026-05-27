package config

import (
	"github.com/spf13/viper"
	"github.com/xmidt-org/sallust"
	"go.uber.org/fx"
)

// Provide requires a *viper.Viper and produces all the configuration
// objects in this package.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			func(v *viper.Viper) (cfg Main, err error) {
				err = v.UnmarshalExact(&cfg)
				return
			},
			func(m Main) DNS {
				return m.DNS
			},
			func(m Main) Groups {
				return m.Groups
			},
			func(m Main) sallust.Config {
				return m.Logging
			},
			func(d DNS) Zone {
				return d.Zone
			},
			func(d DNS) UDPServers {
				return d.UDP
			},
			func(d DNS) TCPServers {
				return d.TCP
			},
		),
	)
}
