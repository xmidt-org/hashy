package server

import (
	"context"
	"time"

	"codeberg.org/miekg/dns"
	"go.uber.org/zap"
)

func NewServerLogger(base *zap.Logger, serverName string) *zap.Logger {
	return base.With(zap.String("server", serverName))
}

// Middleware is a decorator for DNS handlers.
type Middleware struct {
	// Logger is the base logger for hashy.
	Logger *zap.Logger

	// Handler is the hashy Handler that implements the core logic.
	Handler *Handler
}

// decorator wraps a hashy Handler and provides basic middleware functionality.
type decorator struct {
	logger *zap.Logger
	next   *Handler
}

func (d *decorator) ServeDNS(ctx context.Context, response dns.ResponseWriter, request *dns.Msg) {
	start := time.Now()
	logger := d.logger.With(
		zap.Stringer("localAddress", response.LocalAddr()),
		zap.Uint16("id", request.MsgHeader.ID),
		// the request isn't unpacked yet, so we can't see the questions
		// we don't want to unpack it here in case the next handler rejects it without parsing
	)

	defer func() {
		logger.Info("request finished", zap.Duration("duration", time.Since(start)))
	}()

	logger.Info("received request")
	d.next.ServeDNS(ctx, logger, response, request)
}

// Then creates a dns.Handler for a particular dns.Server. The server's logger is also returned,
// so that it can be used in other places, e.g. lifecycle.
func (m *Middleware) Then(serverName string) (handler dns.Handler, serverLogger *zap.Logger) {
	serverLogger = NewServerLogger(m.Logger, serverName)
	handler = &decorator{
		logger: serverLogger,
		next:   m.Handler,
	}

	return
}
