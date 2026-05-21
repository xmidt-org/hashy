// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"errors"

	"github.com/alecthomas/kong"
	"github.com/spf13/viper"
	"github.com/xmidt-org/hashy"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// CommandLine represents the hashy command line.
type CommandLine struct {
	DevMode  bool   `name:"dev" default:"false" help:"developer mode, a standard configuration useful for trying out hashy"`
	ConfFile string `name:"conf-file" help:"configuration file to read. If unset, /etc/hashy, $HOME/.hashy, and the current directory will be searched for hashy.yaml"`
}

func (cl *CommandLine) newViper() (v *viper.Viper, err error) {
	v = viper.New()
	if len(cl.ConfFile) > 0 {
		v.SetConfigFile(cl.ConfFile)
	} else {
		v.SetConfigType("yaml")
		v.SetConfigName("hashy")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.hashy")
		v.AddConfigPath("/etc/hashy")
	}

	if err = v.ReadInConfig(); err == nil || len(cl.ConfFile) > 0 {
		return
	}

	// if a custom config file was specified but not found, we don't load the default config.
	if _, notFound := errors.AsType[viper.ConfigFileNotFoundError](err); notFound {
		// load the default configuration
		v.SetConfigType("yaml")
		err = v.ReadConfig(bytes.NewBufferString(defaultConfig))
	}

	return
}

// AfterApply sets up bindings for Run. Messages from these components are much easier to
// read and debug when done outside an fx.App.
func (cl *CommandLine) AfterApply(ctx *kong.Context) (err error) {
	v, err := cl.newViper()
	if err == nil {
		ctx.Bind(v)
	}

	return
}

func (cl *CommandLine) provideConfig(v *viper.Viper) (config Config, err error) {
	err = v.UnmarshalExact(&config)
	if err == nil {
		if len(config.Zone.Domain) == 0 {
			config.Zone.Domain = DefaultDomain
		}

		if len(config.Groups.DiscoveryDomain) == 0 {
			config.Groups.DiscoveryDomain = DefaultDiscoveryDomain
		}

		if len(config.Groups.GeneratedNamePrefix) == 0 {
			config.Groups.GeneratedNamePrefix = DefaultGeneratedNamePrefix
		}
	}

	return
}

func (cl *CommandLine) provideZoneConfig(config Config) ZoneConfig {
	return config.Zone
}

func (cl *CommandLine) provideGroupsConfig(config Config) GroupsConfig {
	return config.Groups
}

func (cl *CommandLine) provideLogger(_ Config) (*zap.Logger, error) {
	return zap.NewDevelopment() // TODO
}

func (cl *CommandLine) withLogger(l *zap.Logger) fxevent.Logger {
	return &fxevent.ZapLogger{Logger: l}
}

func (cl *CommandLine) provideServerNameGenerator(zcfg ZoneConfig, gcfg GroupsConfig) *hashy.ServerNameGenerator {
	return hashy.NewServerNameGenerator(
		gcfg.GeneratedNamePrefix,
		zcfg.Domain,
	)
}

func (cl *CommandLine) provideFileIngester(nameGenerator *hashy.ServerNameGenerator, logger *zap.Logger, gcfg GroupsConfig) hashy.Ingester {
	fi := &hashy.FileIngester{
		Logger:          logger,
		Globs:           gcfg.ZoneFiles,
		DiscoveryDomain: gcfg.DiscoveryDomain,
		NameGenerator:   nameGenerator,
	}

	return fi
}

// Run executes the hashy server.
func (cl *CommandLine) Run(v *viper.Viper) error {
	app := fx.New(
		fx.Supply(v),
		fx.WithLogger(cl.withLogger),
		fx.Provide(
			cl.provideConfig,
			cl.provideZoneConfig,
			cl.provideGroupsConfig,
			cl.provideLogger,
			cl.provideServerNameGenerator,
			cl.provideFileIngester,
		),
		fx.Invoke(
			func(v *viper.Viper, l *zap.Logger) {
				l.Info("configuration file used", zap.String("location", v.ConfigFileUsed()))
			},
			func(ing hashy.Ingester) {
				ing.Ingest(context.Background())
			},
		),
	)

	app.Run()
	return app.Err()
}
