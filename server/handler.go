// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"io"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy/hashyzap"
	"go.uber.org/zap"
)

const (
	// DefaultZoneDomain is the default DNS domain that hashy serves.
	DefaultZoneDomain = "hashy.net"

	// EndpointLabel defines the subdomain of the zone domain that handles endpoint hashes.
	EndpointLabel = "endpoint"

	// GroupLabel defines the subdomain of the zone domain that handles group DNS lookups.
	GroupLabel = "group"

	// DefaultZoneTTL is the default time-to-live for records generated within
	// hashy's zone.
	DefaultZoneTTL time.Duration = 5 * time.Minute
)

type HandlerOption interface {
	applyToHandler(*Handler) error
}

type handlerOptionFunc func(*Handler) error

func (f handlerOptionFunc) applyToHandler(h *Handler) error { return f(h) }

func WithLogger(base *zap.Logger) HandlerOption {
	return handlerOptionFunc(func(h *Handler) error {
		h.logger = base
		return nil
	})
}

func WithZoneDomain(d string) HandlerOption {
	return handlerOptionFunc(func(h *Handler) error {
		h.zoneDomain = dnsutil.Fqdn(d)
		return nil
	})
}

func WithEndpointHandler(eh *EndpointHandler) HandlerOption {
	return handlerOptionFunc(func(h *Handler) error {
		h.endpointHandler = eh
		return nil
	})
}

func WithGroupHandler(gh *GroupHandler) HandlerOption {
	return handlerOptionFunc(func(h *Handler) error {
		h.groupHandler = gh
		return nil
	})
}

// operation holds all the extracted state necessary for a single Handler request.
type operation struct {
	ctx    context.Context
	writer dns.ResponseWriter

	original *dns.Msg
	start    time.Time

	logger   *zap.Logger
	response *dns.Msg
}

// startOperation initializes a new operation from a DNS request.
func startOperation(ctx context.Context, base *zap.Logger, writer dns.ResponseWriter, original *dns.Msg) (op operation) {
	op.ctx = ctx
	op.writer = writer
	op.original = original
	op.start = time.Now()

	op.response = op.original.Copy()
	op.response.Rcode = dns.RcodeSuccess // default
	dnsutil.SetReply(op.response, op.original)

	op.logger = base.With(
		hashyzap.Request("request", op.original),
	)

	op.logger.Info("request start")
	return
}

// getQuestion attempts to extract the question from the operation's request.
// If this method returns nil, the operation should be abandoned.
func (op *operation) getQuestion() (question dns.RR) {
	switch {
	case len(op.original.Question) != 1:
		op.response.Rcode = dns.RcodeRefused
		op.logger.Error("invalid number of questions")

	case op.original.Question[0].Header().Class != dns.ClassINET:
		op.response.Rcode = dns.RcodeRefused
		op.logger.Error("unhandled class")

	default:
		question = op.original.Question[0]
	}

	return
}

// unhandled reports this operation as unhandled and indicates this in the response.
func (op *operation) unhandled() {
	op.logger.Error("unhandled request")
	op.response.Rcode = dns.RcodeRefused
}

// finish performs all the necessary completion tasks for an operation.
func (op *operation) finish() {
	var err error
	if err = op.response.Pack(); err != nil {
		op.logger.Error("unable to pack response", zap.Error(err))
	}

	if err == nil {
		if _, err = io.Copy(op.writer, op.response); err != nil {
			op.logger.Error("unable to write response", zap.Error(err))
		}
	}

	op.logger.Info("request complete", zap.Duration("duration", time.Since(op.start)))
}

// Handler is the main DNS handler for hashy. Most of hashy's logic is contained
// in this type.
//
// A Handler routes to some internal handlers based on domain.
type Handler struct {
	// logger is the logger for this Handler, typically enhanced with server information.
	logger *zap.Logger

	// zoneDomain is the domain this Handler serves.
	zoneDomain string

	// endpointDomain is the subdomain the endpoint handler serves.
	endpointDomain string

	// endpointHandler is the dns.Handler that serves hashed responses.
	endpointHandler *EndpointHandler

	// groupDomain is the subdomain the group handler serves.
	groupDomain string

	// groupHandler is the dns.Handler servers metadata about groups.
	groupHandler *GroupHandler
}

// NewHandler creates a Handler from a set of options.
func NewHandler(opts ...HandlerOption) (*Handler, error) {
	h := new(Handler)
	for _, o := range opts {
		if err := o.applyToHandler(h); err != nil {
			return nil, err
		}
	}

	if h.endpointHandler == nil {
		return nil, errors.New("an endpoint handler is required")
	}

	if h.groupHandler == nil {
		return nil, errors.New("a group handler is required")
	}

	if h.logger == nil {
		return nil, errors.New("a base logger is required")
	}

	if len(h.zoneDomain) == 0 {
		h.zoneDomain = DefaultZoneDomain
	}

	h.endpointDomain = dnsutil.Join(EndpointLabel, h.zoneDomain)
	h.groupDomain = dnsutil.Join(GroupLabel, h.zoneDomain)

	return h, nil
}

// Clone creates a copy of this handler that uses the given logger, which is
// typically a server-specific logger.
//
// If logger is nil, zap.NewNop() is used.
func (h *Handler) Clone(logger *zap.Logger) *Handler {
	clone := new(Handler)
	*clone = *h
	clone.logger = logger
	if clone.logger == nil {
		clone.logger = zap.NewNop()
	}

	return clone
}

func (h *Handler) ServeDNS(ctx context.Context, writer dns.ResponseWriter, request *dns.Msg) {
	op := startOperation(ctx, h.logger, writer, request)
	defer op.finish()

	question := op.getQuestion()
	if question == nil {
		return
	}

	switch {
	case dnsutil.IsBelow(h.endpointDomain, question.Header().Name):
		h.endpointHandler.ServeRequest(
			op.ctx,
			op.logger,
			op.response,
			ParseEndpointRequest(question, h.endpointDomain),
		)

	case dnsutil.IsBelow(h.groupDomain, question.Header().Name):
		h.groupHandler.ServeRequest(
			op.ctx,
			op.logger,
			op.response,
			ParseGroupRequest(question, h.groupDomain),
		)

	default:
		op.unhandled()
	}
}
