package server

import (
	"context"

	"codeberg.org/miekg/dns"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Lifecycler handles binding servers to the fx.App lifecycle.
type Lifecycler struct {
	Lifecycle  fx.Lifecycle
	Shutdowner fx.Shutdowner
}

// start returns is the goroutine that runs a *dns.Server.
func (lc *Lifecycler) start(serverLogger *zap.Logger, server *dns.Server) {
	// anytime any server shuts down, stop the app
	defer lc.Shutdowner.Shutdown()
	serverLogger.Info("starting server")

	// ListenAndServe returns a nil error when it terminates normally,
	// unlike net/http.Server.
	if err := server.ListenAndServe(); err != nil {
		serverLogger.Error("unable to start server", zap.Error(err))
	} else {
		serverLogger.Info("server stopped")
	}
}

// startHook returns an fx.HookFunc that starts the given *dns.Server.
func (lc *Lifecycler) startHook(serverLogger *zap.Logger, server *dns.Server) func() {
	return func() {
		go lc.start(serverLogger, server)
	}
}

// stopHook returns an fx.HookFunc that shuts down the given *dns.Server.
func (lc *Lifecycler) stopHook(server *dns.Server) func(context.Context) {
	return server.Shutdown
}

// Append binds the server to the enclosing fx.App lifecycle.
func (lc *Lifecycler) Append(serverLogger *zap.Logger, server *dns.Server) {
	lc.Lifecycle.Append(
		fx.StartStopHook(
			lc.startHook(serverLogger, server),
			lc.stopHook(server),
		),
	)
}
