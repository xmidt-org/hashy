package main

import (
	"github.com/spf13/viper"
	"github.com/xmidt-org/hashy"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// TODO: drive logging from the configuration
func newLogger(v *viper.Viper, _ hashy.Config) (l *zap.Logger, err error) {
	l, err = zap.NewDevelopment()
	if err == nil {
		l.Info("configuration file", zap.String("path", v.ConfigFileUsed()))
	}

	return
}

func withLogger(l *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{
		Logger: l,
	}
}

func provideLogging() fx.Option {
	return fx.Options(
		fx.Provide(
			newLogger,
		),
		fx.WithLogger(
			withLogger,
		),
	)
}
