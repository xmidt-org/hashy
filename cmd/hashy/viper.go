package main

import (
	"errors"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

func createViper() (v *viper.Viper, err error) {
	v = viper.New()
	v.SetConfigName("hashy")

	// POSIX-style search paths
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME/.hashy")
	v.AddConfigPath("/etc/hashy")

	err = v.ReadInConfig()

	var notFoundErr viper.ConfigFileNotFoundError
	if errors.As(err, &notFoundErr) {
		// TODO: ignore for now
		err = nil
	}

	return
}

func provideViper() fx.Option {
	return fx.Provide(
		createViper,
	)
}
