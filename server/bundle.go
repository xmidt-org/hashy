package server

import (
	"context"
	"fmt"

	"codeberg.org/miekg/dns"
	"github.com/xmidt-org/hashy/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Info holds the components for a single DNS server.
type Info struct {
	Name   string
	Server *dns.Server
	Logger *zap.Logger
}

// Start is the lifecycle hook that starts the server referred to by this Info.
// If the server couldn't start, or the server stops, the supplied shutdowner is
// used to shutdown the enclosing fx.App.
func (i Info) Start(sh fx.Shutdowner) (err error) {
	defer sh.Shutdown()
	i.Logger.Info("starting server")

	if err = i.Server.ListenAndServe(); err == nil {
		// ListenAndServe returns a nil error when it terminates normally,
		// unlike net/http.Server which returns a non-nil error.
		i.Logger.Info("server stopped")
	} else {
		i.Logger.Error("unable to start server", zap.Error(err))
	}

	return
}

// Stop is the lifecycle hook that stops the server referred to by this Info.
func (i Info) Stop(ctx context.Context) {
	i.Server.Shutdown(ctx)
}

// Bundle holds Info objects keyed by their name.
type Bundle map[string]Info

// Add adds a *dns.Server to this Bundle. If the given server is a duplicate,
// this method returns an error. This method also creates this Bundle as needed.
func (m *Bundle) Add(name string, server *dns.Server) error {
	info := Info{
		Name:   name,
		Server: server,
	}

	if *m == nil {
		*m = Bundle{
			info.Name: info,
		}

		return nil
	}

	if _, exists := (*m)[info.Name]; exists {
		return fmt.Errorf("duplicate server name: %s", info.Name)
	}

	(*m)[info.Name] = info
	return nil
}

// UseLogger establishes a sublogger for every server in this Bundle using
// the supplied parent logger.
func (m Bundle) UseLogger(parent *zap.Logger) {
	for name, info := range m {
		info.Logger = NewServerLogger(parent, name, info.Server)
		m[name] = info
	}
}

// UseHandler clones the given handler for each server, configuring each
// clone with the server logger.
//
// The DNS package doesn't allow setting anything in the context, so this method
// handles server-specific logging in handlers.
func (m Bundle) UseHandler(base *Handler) {
	for name, info := range m {
		info.Server.Handler = base.Clone(info.Logger)
		m[name] = info
	}
}

// BindToLifecycle attaches all servers in this Bundle to the enclosing fx.App lifecycle.
// If any server's exit, the given shutdowner is used to shutdown the entire app.
//
// The logger supplied via UseLogger is used in the lifecycle hooks.
func (m Bundle) BindToLifecycle(lc fx.Lifecycle, sh fx.Shutdowner) {
	for _, info := range m {
		lc.Append(
			fx.StartStopHook(
				func() {
					go info.Start(sh)
				},
				info.Stop,
			),
		)
	}
}

// NewBundle creates all the servers from configuration and returns a Bundle
// containing them.
func NewBundle(cfg config.DNS, parent *zap.Logger) (servers Bundle, err error) {
	servers = make(Bundle, len(cfg.UDP)+len(cfg.TCP))
	for name, udpConfig := range cfg.UDP {
		var server *dns.Server
		if server, err = NewUDPServer(udpConfig); err == nil {
			err = servers.Add(name, server)
		}

		if err != nil {
			return
		}
	}

	for name, tcpConfig := range cfg.TCP {
		var server *dns.Server
		if server, err = NewTCPServer(tcpConfig); err == nil {
			err = servers.Add(name, server)
		}

		if err != nil {
			return
		}
	}

	servers.UseLogger(parent)
	return
}
