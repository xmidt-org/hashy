// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"

	"github.com/alecthomas/kong"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// CommandLine represents the hashy command line.
type CommandLine struct {
	DevMode  bool   `name:"dev" default:"false" help:"developer mode, a standard configuration useful for trying out hashy"`
	ConfFile string `name:"conf-file" help:"configuration file to read. If unset, /etc/hashy, $HOME/.hashy, and the current directory will be searched for hashy.yaml"`
}

func (cl *CommandLine) newViper() (v *viper.Viper, loc ConfigLocation, err error) {
	v = viper.New()
	if len(cl.ConfFile) > 0 {
		v.SetConfigFile(cl.ConfFile)
		loc = ConfigLocation(cl.ConfFile)
	} else {
		v.SetConfigType("yaml")
		v.SetConfigName("hashy")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
		v.AddConfigPath("/etc/hashy")
	}

	err = v.ReadInConfig()
	switch {
	case err == nil:
		loc = ConfigLocation(v.ConfigFileUsed())

	case len(cl.ConfFile) == 0:
		// if a custom config file was specified but not found, we don't load the default config.
		if _, notFound := errors.AsType[viper.ConfigFileNotFoundError](err); notFound {
			// load the default configuration
			v.SetConfigType("yaml")
			err = v.ReadConfig(bytes.NewBufferString(defaultConfig))
			if err == nil {
				loc = ConfigLocation("default")
			}
		}
	}

	return
}

// AfterApply sets up bindings for Run. Messages from these components are much easier to
// read and debug when done outside an fx.App.
func (cl *CommandLine) AfterApply(ctx *kong.Context) (err error) {
	v, loc, err := cl.newViper()
	if err == nil {
		ctx.Bind(v, loc)
	}

	return
}

func (cl *CommandLine) provideConfig(v *viper.Viper, loc ConfigLocation) (config Config, err error) {
	err = v.UnmarshalExact(&config)
	return
}

func (cl *CommandLine) providerLogger(_ Config) (*zap.Logger, error) {
	return zap.NewDevelopment() // TODO
}

func (cl *CommandLine) withLogger(l *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: l}
}

// Run executes the hashy server.
func (cl *CommandLine) Run(v *viper.Viper, loc ConfigLocation) error {
	app := fx.New(
		fx.Supply(v, loc),
		fx.WithLogger(cl.withLogger),
		fx.Provide(
			cl.provideConfig,
			cl.providerLogger,
		),
		fx.Invoke(
			func(l *zap.Logger) {
				l.Info("configuration file used", zap.String("location", string(loc)))
			},
		),
	)

	app.Run()
	return app.Err()
}
