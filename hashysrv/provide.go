// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashysrv

import (
	"github.com/miekg/dns"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Provide builds the necessary components to start all the hashy server instances.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			NewHandler,
			NewServers,
		),
		fx.Invoke(
			// ensure the dependency graph gets built
			func(l *zap.Logger, _ []*dns.Server) {
				l.Info("all hashy servers started")
			},
		),
	)
}
