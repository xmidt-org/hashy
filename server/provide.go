// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/config"
	"github.com/xmidt-org/hashy/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Provide builds the server components.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			func(zcfg config.Zone, l *zap.Logger) (j *hashy.TTLJitterer, err error) {
				j, err = hashy.NewTTLJitterer(
					hashy.DurationToSeconds(zcfg.TTL),
					zcfg.TTLJitter,
				)

				if err == nil {
					lo, hi := j.Range()
					l.Debug("TTL jitter range for generated DNS RRs",
						zap.Uint32("lo", lo), zap.Uint32("hi", hi))
				}

				return
			},
			// create the base handler that will be cloned for each server
			func(zcfg config.Zone, j *hashy.TTLJitterer, loc *service.Locator) (*Handler, error) {
				return NewHandler(
					WithZoneDomain(zcfg.Domain),
					WithJitterer(j),
					WithLocator(loc),
				)
			},

			// create the server Bundle and bind it to the fx.App lifecycle
			func(dcfg config.DNS, parent *zap.Logger, base *Handler, lc fx.Lifecycle, sh fx.Shutdowner) (b Bundle, err error) {
				b, err = NewBundle(dcfg, parent)
				if err == nil {
					b.UseHandler(base)
					b.BindToLifecycle(lc, sh)
				}

				return
			},
		),
		fx.Invoke(
			// ensure that the dns servers are always created
			func(Bundle) {},
		),
	)
}
