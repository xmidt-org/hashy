package hashysrv

import (
	"context"

	"github.com/miekg/dns"
	"github.com/xmidt-org/hashy/hashycfg"
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
	Handler dns.Handler
	Config  hashycfg.Config
}

// NewServers creates the configured hashy servers and binds them to
// the enclosing app lifecycle.
func NewServers(in ServerIn) (servers []*dns.Server) {
	servers = in.Config.Servers.NewServers()
	if len(servers) == 0 {
		// ensure some defaults
		servers = append(servers, hashycfg.Server{}.NewServer())
		servers = append(servers,
			hashycfg.Server{
				Network: "tcp",
			}.NewServer(),
		)
	}

	for _, s := range servers {
		s.Handler = in.Handler
		in.Lifecycle.Append(
			fx.Hook{
				OnStart: OnStartServer(in, s),
				OnStop:  OnStopServer(in, s),
			},
		)
	}

	return
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

// Provide builds the necessary components to start all the hashy server instances.
func Provide() fx.Option {
	return fx.Options(
		fx.Provide(
			NewHandler,
			NewServers,
		),
		fx.Invoke(
			// ensure the dependency graph gets built
			func(l *zap.Logger, _ []*dns.Server) {
				l.Info("all hashy servers started")
			},
		),
	)
}
