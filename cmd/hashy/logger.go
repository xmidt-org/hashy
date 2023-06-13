package main

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func newFXEventLogger(l *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{
		Logger: l,
	}
}

func provideLogger() fx.Option {
	return fx.Options(
		fx.WithLogger(newFXEventLogger),
		fx.Provide(
			zap.NewDevelopment,
		),
	)
}
