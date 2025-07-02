// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/xmidt-org/hashy/hashycfg"
	"github.com/xmidt-org/hashy/hashysrv"
	"github.com/xmidt-org/hashy/hashyzap"
	"go.uber.org/fx"
)

const (
	hashyDescription = "hashy is a DNS server that uses a consistent hash for certain domains and passes through others"
)

// CommandLine represents the hashy command line
type CommandLine struct {
	DevMode bool `name:"dev" default:"false" help:"developer mode, a standard configuration useful for trying out hashy"`
}

// run parses the command line, builds an fx.App, and runs the application.
// args is expected to be only the program arguments, e.g. os.Args[1:].
//
// The kong options can be used to set options for an embedded execution,
// such as trapping stdout and stderr for unit tests.
func run(args []string, opts ...kong.Option) {
	// always ensure certain options are set
	opts = append(opts,
		kong.Description(hashyDescription),
	)

	var ctx *kong.Context
	cli := new(CommandLine)
	k, err := kong.New(cli, opts...)
	if err == nil {
		ctx, err = k.Parse(args)
	}

	if err != nil {
		// usage and the error have already been printed
		if k != nil {
			k.Exit(1)
		}

		return
	}

	app := fx.New(
		fx.Supply(cli),
		fx.Supply(ctx),
		hashycfg.Provide(),
		hashyzap.Provide(),
		hashysrv.Provide(),
	)

	app.Run()
	if err := app.Err(); err != nil {
		fmt.Fprintf(
			k.Stderr,
			"Unable to start hashy: %s\n",
			err,
		)

		k.Exit(2)
	}
}
