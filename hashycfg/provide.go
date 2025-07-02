// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashycfg

import "go.uber.org/fx"

// Provide establishes the configuration, possibly unmarshaled externally,
// for a hashy process.
func Provide() fx.Option {
	return fx.Provide(
		NewViper,
		Unmarshal,
	)
}
