package server

import (
	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy/config"
	"github.com/xmidt-org/hashy/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Provide builds the server components.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			func(z config.Zone, loc *service.Locator) *Handler {
				h := &Handler{
					Domain:  z.Domain,
					Locator: loc,
				}

				if len(h.Domain) == 0 {
					h.Domain = DefaultZoneDomain
				}

				h.Domain = dnsutil.Fqdn(h.Domain)
				return h
			},
			func(base *zap.Logger, h *Handler) *Middleware {
				return &Middleware{
					Logger:  base,
					Handler: h,
				}
			},
			func(l fx.Lifecycle, sh fx.Shutdowner) *Lifecycler {
				return &Lifecycler{
					Lifecycle:  l,
					Shutdowner: sh,
				}
			},
			NewDNSServers,
		),
		fx.Invoke(
			// ensure that the dns servers are always created
			func([]*dns.Server) {},
		),
	)
}
