// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

const (
	hashyDescription = "hashy is a DNS server that uses a consistent hash for certain domains and passes through others"
)

// run parses the command line, builds an fx.App, and runs the application.
// args is expected to be only the program arguments, e.g. os.Args[1:].
//
// The kong options can be used to set options for an embedded execution,
// such as trapping stdout and stderr for unit tests.
func run(args []string, opts ...kong.Option) (err error) {
	// always ensure certain options are set
	opts = append(opts,
		kong.Description(hashyDescription),
	)

	var cli CommandLine
	k, err := kong.New(&cli, opts...)
	if err != nil {
		err = fmt.Errorf("error creating command line parser: %w", err)
		return
	}

	ctx, err := k.Parse(args)
	if err != nil {
		err = fmt.Errorf("error parsing command line: %w", err)
		return
	}

	err = ctx.Run()
	return
}
