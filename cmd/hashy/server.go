package main

import (
	"github.com/miekg/dns"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func provideServer() fx.Option {
	return fx.Provide(
		func(l *zap.Logger, lf fx.Lifecycle, sh fx.Shutdowner) (s *dns.Server, err error) {
			s = &dns.Server{
				Addr: ":5353",
			}

			lf.Append(
				fx.StartStopHook(
					func() {
						go func() {
							defer sh.Shutdown()
							err := s.ListenAndServe()
							if err != nil {
								l.Error("aborted", zap.Error(err))
							}
						}()
					},
					s.ShutdownContext,
				),
			)

			return
		},
	)
}
