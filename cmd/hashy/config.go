package main

import (
	"errors"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	Servers []string `json:"servers" yaml:"servers"`
}

func newViper() (v *viper.Viper, err error) {
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
}

func newConfig(v *viper.Viper) (cfg Config, err error) {
	err = v.UnmarshalExact(&cfg)
	return
}

func provideConfig() fx.Option {
	return fx.Provide(
		newViper,
		newConfig,
	)
}
