package server

import (
	"github.com/xmidt-org/hashy/config"
	"github.com/xmidt-org/hashy/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Provide builds the server components.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			// create the base handler that will be cloned for each server
			func(zcfg config.Zone, loc *service.Locator) (*Handler, error) {
				return NewHandler(
					WithZoneConfig(zcfg),
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
