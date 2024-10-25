package main

import (
	"github.com/xmidt-org/hashy"
	"go.uber.org/fx"
)

func provideHandler() fx.Option {
	return fx.Provide(
		hashy.NewHandler,
	)
}
