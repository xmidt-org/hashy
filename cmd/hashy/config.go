package main

import (
	"errors"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	Servers []string `json:"servers" yaml:"servers"`
}

func provideConfig() fx.Option {
	return fx.Provide(
		func() (v *viper.Viper, err error) {
			v = viper.New()
			v.AddConfigPath("/etc/hashy")
			v.AddConfigPath("$HOME/.hashy")
			v.AddConfigPath(".")
			v.SetConfigName("hashy")

			err = v.ReadInConfig()

			var notFoundErr viper.ConfigFileNotFoundError
			if errors.As(notFoundErr, &notFoundErr) {
				err = nil // ignore not found errors, and fallback to defaults
			}

			return
		},
		func(v *viper.Viper) (cfg Config, err error) {
			err = v.UnmarshalExact(&cfg)
			return
		},
	)
}
