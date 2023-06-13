package main

import (
	"github.com/miekg/dns"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Handler struct {
	Logger *zap.Logger
}

func (h *Handler) ServeDNS(rw dns.ResponseWriter, request *dns.Msg) {
	h.Logger.Info("received request")
	response := new(dns.Msg)
	response.SetReply(request)
	response.Answer = []dns.RR{
		&dns.SRV{
			Hdr: dns.RR_Header{
				Ttl: 3600,
			},
			Port:   8080,
			Target: "talaria-xyz.xmidt.comcast.net",
		},
	}

	rw.WriteMsg(response)
	rw.Close()

}

func newHandler(l *zap.Logger) *Handler {
	return &Handler{
		Logger: l,
	}
}

func provideHandler() fx.Option {
	return fx.Provide(
		newHandler,
	)
}
