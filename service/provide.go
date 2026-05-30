// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"

	"github.com/xmidt-org/hashy/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Provide creates the relevant components in this package.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func(base *zap.Logger, gcfg config.Groups) (loc *Locator, lis IngestListener, err error) {
					loc, err = NewLocator(
						WithLocatorLogger(base),
						WithVNodes(gcfg.VNodes),
					)

					if err == nil {
						lis = loc
					}

					return
				},
				fx.ResultTags("", `group:"ingestListeners"`),
			),
			fx.Annotate(
				func(base *zap.Logger, gcfg config.Groups, listeners []IngestListener) (*FileIngester, error) {
					return NewFileIngester(
						WithIngestLogger(base),
						WithGroupsConfig(gcfg),
						WithIngestListeners(listeners...),
					)
				},
				fx.ParamTags("", "", `group:"ingestListeners"`),
			),
			func(gcfg config.Groups, fi *FileIngester, lc fx.Lifecycle) (ic *IngestChecker, err error) {
				ic, err = NewIngestChecker(
					WithIngester(fi),
					WithCheckInterval(gcfg.CheckInterval),
				)

				if err != nil {
					return
				}

				lc.Append(fx.StartStopHook(
					ic.Start,
					ic.Stop,
				))

				return
			},
		),
		fx.Invoke(
			func(fi *FileIngester) {
				fi.Ingest(context.Background())
			},
			func(*IngestChecker) {},
		),
	)
}
