package service

import (
	"context"

	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/config"
	"github.com/xmidt-org/medley/consistent"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Provide creates the relevant components in this package.
//
// The following components must be supplied:
//
//   - *zap.Logger
//   - config.Groups
//   - config.Zone
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			func(gcfg config.Groups, zcfg config.Zone) EndpointNameGenerator {
				return EndpointNameGenerator{
					Prefix: gcfg.GeneratedNamePrefix,
					Domain: zcfg.Domain,
				}
			},
			func(base *zap.Logger, eng EndpointNameGenerator, gcfg config.Groups) *FileIngester {
				return &FileIngester{
					Logger:          base.Named("fileIngester"),
					ZoneFiles:       gcfg.ZoneFiles,
					Origin:          gcfg.Origin,
					DefaultTTL:      hashy.DurationToSeconds(gcfg.DefaultTTL),
					DiscoveryDomain: gcfg.DiscoveryDomain,
					NameGenerator:   eng,
				}
			},
			func(base *zap.Logger, fi *FileIngester, gcfg config.Groups) *Locator {
				return &Locator{
					logger: base.Named("locator"),
					builder: new(consistent.Builder[string, *Endpoint]).
						VNodes(gcfg.VNodes),
				}
			},
			fx.Annotate(
				func(loc *Locator) IngestListener {
					return loc
				},
				fx.ResultTags(`group:"ingestListeners"`),
			),
		),
		fx.Invoke(
			fx.Annotate(
				func(fi *FileIngester, listeners []IngestListener) {
					fi.AddIngestListeners(listeners...)
				},
				fx.ParamTags(``, `group:"ingestListeners"`),
			),
			func(fi *FileIngester) {
				fi.Ingest(context.Background())
			},
		),
	)
}
