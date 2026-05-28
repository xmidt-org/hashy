package server

import (
	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
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
			func(z config.Zone, loc *service.Locator) *Handler {
				h := &Handler{
					zoneDomain: z.Domain,
					locator:    loc,
				}

				if len(h.zoneDomain) == 0 {
					h.zoneDomain = DefaultZoneDomain
				}

				h.groupsDomain = dnsutil.Join(GroupsLabel, h.zoneDomain)

				if z.TTL > 0 {
					h.ttl = hashy.DurationToSeconds(z.TTL)
				} else {
					h.ttl = hashy.DurationToSeconds(DefaultZoneTTL)
				}

				h.zoneDomain = dnsutil.Fqdn(h.zoneDomain)
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
