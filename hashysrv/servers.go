package hashysrv

import (
	"context"

	"github.com/miekg/dns"
	"github.com/xmidt-org/hashy/hashycfg"
	"github.com/xmidt-org/hashy/hashyzap"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// ServerIn is the set of dependencies needed in order to bootstrap the
// hashy server instances.
type ServerIn struct {
	fx.In

	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner

	Logger  *zap.Logger
	Handler *Handler
	Config  hashycfg.Config
}

// OnStartServer creates a lifecycle callback for starting the given server.
// When the server exits for any reason, the Shutdowner is invoked.
func OnStartServer(in ServerIn, s *dns.Server) func(context.Context) error {
	return func(context.Context) error {
		go func() {
			defer in.Shutdowner.Shutdown()
			err := s.ListenAndServe()
			in.Logger.Info("server exit", zap.Error(err))
		}()

		return nil
	}
}

// OnStopServer creates a lifecycle callback for stopping the given server.
func OnStopServer(in ServerIn, s *dns.Server) func(context.Context) error {
	return func(ctx context.Context) error {
		return s.ShutdownContext(ctx)
	}
}

// NewServers creates the configured hashy servers and binds them to
// the enclosing app lifecycle.
func NewServers(in ServerIn) (servers []*dns.Server) {
	servers = in.Config.Servers.NewServers()
	for _, s := range servers {
		// use the configured handler as a prototype
		handler := new(Handler)
		*handler = *in.Handler
		handler.logger = handler.logger.With(hashyzap.ServerField("server", s))
		s.Handler = handler

		in.Lifecycle.Append(
			fx.Hook{
				OnStart: OnStartServer(in, s),
				OnStop:  OnStopServer(in, s),
			},
		)
	}

	return
}
