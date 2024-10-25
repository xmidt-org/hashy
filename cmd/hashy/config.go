package main

import (
	"github.com/spf13/viper"
	"github.com/xmidt-org/hashy"
	"go.uber.org/fx"
)

func unmarshalConfig(v *viper.Viper) (cfg hashy.Config, err error) {
	err = v.UnmarshalExact(&cfg)
	return
}

func provideConfig() fx.Option {
	return fx.Provide(
		unmarshalConfig,
	)
}
