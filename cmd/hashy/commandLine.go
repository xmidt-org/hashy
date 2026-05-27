// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"

	"github.com/alecthomas/kong"
	"github.com/spf13/viper"
	"github.com/xmidt-org/hashy/config"
	"github.com/xmidt-org/hashy/server"
	"github.com/xmidt-org/hashy/service"
	"github.com/xmidt-org/sallust"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// CommandLine represents the hashy command line.
type CommandLine struct {
	ConfFile  string   `name:"conf-file" help:"configuration file to read. If unset, /etc/hashy, $HOME/.hashy, and the current directory will be searched for hashy.yaml"`
	ZoneFiles []string `name:"zone-files" help:"additional globs for zone files. Will be appended to configuration. Relative paths are resolved relative to the conf file."`
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
		err = v.ReadConfig(bytes.NewBufferString(config.Default))
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

func (cl *CommandLine) provideLogging() fx.Option {
	return fx.Options(
		fx.Provide(
			func(cfg sallust.Config) (*zap.Logger, error) {
				return cfg.Build()
			},
		),
		fx.WithLogger(
			func(l *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: l}
			},
		),
	)
}

// decorateGroups adds the command-line zone files, if any, to the ZoneFiles configuration
// value. Additionally, all ZoneFiles are expanded.
func (cl *CommandLine) decorateGroups(v *viper.Viper, gcfg config.Groups) config.Groups {
	gcfg.ZoneFiles = slices.Grow(gcfg.ZoneFiles, len(cl.ZoneFiles))
	gcfg.ZoneFiles = append(gcfg.ZoneFiles, cl.ZoneFiles...)

	configLocation := v.ConfigFileUsed()
	if len(configLocation) > 0 {
		configLocation = filepath.Dir(configLocation)
	}

	for i, glob := range gcfg.ZoneFiles {
		glob = os.ExpandEnv(glob)
		if !filepath.IsAbs(glob) {
			glob = filepath.Join(configLocation, glob)
		}

		gcfg.ZoneFiles[i] = glob
	}

	return gcfg
}

// Run executes the hashy server.
func (cl *CommandLine) Run(v *viper.Viper) error {
	app := fx.New(
		fx.Supply(v),
		cl.provideLogging(),
		config.Provide(),
		service.Provide(),
		server.Provide(),
		fx.Decorate(cl.decorateGroups),
		fx.Invoke(
			func(v *viper.Viper, l *zap.Logger) {
				l.Info("configuration file used", zap.String("location", v.ConfigFileUsed()))
			},
		),
	)

	app.Run()
	return app.Err()
}
