package main

import (
	"context"

	"github.com/miekg/dns"
	"github.com/xmidt-org/hashy"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServerIn struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner

	Logger  *zap.Logger
	Handler dns.Handler
	Config  hashy.Config
}

func newServers(in ServerIn) (servers []*dns.Server) {
	servers = in.Config.Servers.NewServers()
	if len(servers) == 0 {
		// ensure some defaults
		servers = append(servers, hashy.ServerConfig{}.NewServer())
		servers = append(servers,
			hashy.ServerConfig{
				Network: "tcp",
			}.NewServer(),
		)
	}

	for _, s := range servers {
		s.Handler = in.Handler
		in.Lifecycle.Append(
			fx.Hook{
				OnStart: onStartServer(in, s),
				OnStop:  onStopServer(in, s),
			},
		)
	}

	return
}

func onStartServer(in ServerIn, s *dns.Server) func(context.Context) error {
	return func(context.Context) error {
		go func() {
			defer in.Shutdowner.Shutdown()
			err := s.ListenAndServe()
			in.Logger.Info("server exit", zap.Error(err))
		}()

		return nil
	}
}

func onStopServer(in ServerIn, s *dns.Server) func(context.Context) error {
	return func(ctx context.Context) error {
		return s.ShutdownContext(ctx)
	}
}

func provideServers() fx.Option {
	return fx.Options(
		fx.Provide(
			newServers,
		),
		fx.Invoke(
			// ensure the dependency graph gets built
			func([]*dns.Server) {},
		),
	)
}
