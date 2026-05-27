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
			func(l *zap.Logger, eng EndpointNameGenerator, gcfg config.Groups) *FileIngester {
				return &FileIngester{
					Logger:          l,
					ZoneFiles:       gcfg.ZoneFiles,
					Origin:          gcfg.Origin,
					DefaultTTL:      hashy.DurationToSeconds(gcfg.DefaultTTL),
					DiscoveryDomain: gcfg.DiscoveryDomain,
					NameGenerator:   eng,
				}
			},
			func(fi *FileIngester, gcfg config.Groups) *Locator {
				l := &Locator{
					builder: new(consistent.Builder[string, *Endpoint]).
						VNodes(gcfg.VNodes),
				}

				fi.AddIngestListener(l)
				return l
			},
		),
		fx.Invoke(
			func(fi *FileIngester) {
				fi.Ingest(context.Background())
			},
		),
	)
}
