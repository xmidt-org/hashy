package main

import (
	"github.com/miekg/dns"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newDNSServer(l *zap.Logger, h *Handler, lf fx.Lifecycle, sh fx.Shutdowner) (s *dns.Server, err error) {
	s = &dns.Server{
		Addr:    ":1111",
		Net:     "udp",
		Handler: h,
		NotifyStartedFunc: func() {
			l.Info("server started")
		},
	}

	lf.Append(
		fx.StartStopHook(
			func() {
				go func() {
					l.Info("staring server ...")
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
}

func provideServer() fx.Option {
	return fx.Provide(
		newDNSServer,
	)
}
