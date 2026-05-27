package server

import (
	"context"
	"fmt"
	"io"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy/service"
	"github.com/xmidt-org/sallust"
	"go.uber.org/zap"
)

const (
	DefaultZoneDomain = service.DefaultGeneratedEndpointDomain
)

// Handler is the main DNS handler for hashy.
//
// This type does not implement dns.Handler. ServeDNS takes a logger that is enriched by Middleware.
type Handler struct {
	Domain  string
	Locator *service.Locator
}

func (h *Handler) handleQuestion(_ context.Context, logger *zap.Logger, response *dns.Msg, question dns.RR) {
	logger.Info("received question", zap.Stringer("question", question))
	if question.Header().Name != h.Domain {
		response.Rcode = dns.RcodeNameError
		return
	}

	// TODO: just refuse everything for now
	response.Rcode = dns.RcodeRefused
}

func (h *Handler) ServeDNS(ctx context.Context, logger *zap.Logger, writer dns.ResponseWriter, request *dns.Msg) {
	response := request.Copy()
	dnsutil.SetReply(response, request)

	request.Unpack()
	if len(request.Question) == 1 {
		question := request.Question[0]
		logger := sallust.Get(ctx).With(
			zap.Stringer("question", fmt.Stringer(question)),
		)

		h.handleQuestion(ctx, logger, response, request.Question[0])
	} else {
		response.Rcode = dns.RcodeRefused
	}

	response.Pack()
	io.Copy(writer, response)
}
