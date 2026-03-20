// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashyzap

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// TODO: drive logging from the configuration
func newLogger(v *viper.Viper) (l *zap.Logger, err error) {
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

// Provide sets up a zap.Logger for both fx and the application layer.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			newLogger,
		),
		fx.WithLogger(
			withLogger,
		),
	)
}
