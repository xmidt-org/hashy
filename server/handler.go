// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"io"
	"iter"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/hashyzap"
	"github.com/xmidt-org/hashy/service"
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

// groupRequest holds the information from a DNS question for responding with located endpoints.
type endpointRequest struct {
	name   string
	groups []string

	prefix string
	object string
	extra  []string

	rrType uint16
}

// groupRequest holds the information from a DNS question for responding with group metadata.
type groupRequest struct {
	name string

	group  string
	rrType uint16
}

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

func WithZoneDomain(zoneDomain string) HandlerOption {
	return handlerOptionFunc(func(h *Handler) error {
		if len(zoneDomain) > 0 {
			h.zoneDomain = dnsutil.Fqdn(zoneDomain)
			h.endpointDomain = dnsutil.Join(EndpointLabel, h.zoneDomain)
			h.groupDomain = dnsutil.Join(GroupLabel, h.zoneDomain)
		}

		return nil
	})
}

func WithLocator(l *service.Locator) HandlerOption {
	return handlerOptionFunc(func(h *Handler) error {
		h.locator = l
		return nil
	})
}

func WithJitterer(j *hashy.TTLJitterer) HandlerOption {
	return handlerOptionFunc(func(h *Handler) (err error) {
		h.jitterer = j
		return
	})
}

// Handler is the main DNS handler for hashy. Most of hashy's logic is contained
// in this type.
type Handler struct {
	// logger is the logger for this Handler, typically enhanced with server information.
	logger *zap.Logger

	// zoneDomain is the domain that hashy serves.
	zoneDomain string

	// endpointDomain is the subdomain that serves endpoint hashes.
	endpointDomain string

	// groupDomain is the subdomain of zoneDomain that hashy responds to with group metadata.
	groupDomain string

	// locator is the required service locator for this handler.
	locator *service.Locator

	// jitterer is used to provide jitter for the TTL of any DNS RRs produced by this handler.
	jitterer *hashy.TTLJitterer
}

// NewHandler creates a Handler from a set of options.
func NewHandler(opts ...HandlerOption) (*Handler, error) {
	h := new(Handler)
	for _, o := range opts {
		if err := o.applyToHandler(h); err != nil {
			return nil, err
		}
	}

	if h.locator == nil {
		return nil, fmt.Errorf("a locator is required for a Handler")
	}

	if h.logger == nil {
		h.logger = zap.NewNop()
	}

	if len(h.zoneDomain) == 0 {
		WithZoneDomain(DefaultZoneDomain).applyToHandler(h)
	}

	if h.jitterer == nil {
		h.jitterer, _ = hashy.NewTTLJitterer(hashy.DurationToSeconds(DefaultZoneTTL), 0.0)
	}

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

func (h *Handler) writeResponse(logger *zap.Logger, writer dns.ResponseWriter, response *dns.Msg) {
	if err := response.Pack(); err != nil {
		logger.Error("unable to pack response", zap.Error(err))
		return
	}

	if _, err := io.Copy(writer, response); err != nil {
		logger.Error("unable to write response", zap.Error(err))
	}
}

// parseEndpointRequest parses the question into a request object carrying the
// information to satisfy and endpoint hash.
func (h *Handler) parseEndpointRequest(question dns.RR) endpointRequest {
	var (
		request = endpointRequest{
			name:   question.Header().Name,
			rrType: dns.RRToType(question),
		}

		subdomain = dnsutil.Trim(request.name, h.endpointDomain)
		labels    = strings.Split(subdomain, ".")

		// use only the first (leftmost) label to extract the hash object
		parts = strings.Split(labels[0], "-")
	)

	// the labels between the hash object and the zone domain are
	// taken to be group names used as filters
	request.groups = labels[1:]

	if len(parts) > 1 {
		request.prefix = parts[0]
		request.object = parts[1]
		request.extra = parts[2:]
	} else {
		request.object = parts[0]
	}

	return request
}

func (h *Handler) serveEndpoint(response *dns.Msg, request endpointRequest) {
	endpoints := h.locator.FindString(request.object, request.groups...)
	response.Answer = slices.Grow(response.Answer, endpoints.LenRRs(request.rrType))

	header := dns.Header{
		Name:  request.name,
		TTL:   h.jitterer.TTL(),
		Class: dns.ClassINET,
	}

	for _, rr := range endpoints.RRs(request.rrType) {
		*rr.Header() = header
		response.Answer = append(response.Answer, rr)
	}

	if len(response.Answer) > 1 {
		rand.Shuffle(len(response.Answer), func(i, j int) {
			response.Answer[i], response.Answer[j] =
				response.Answer[j], response.Answer[i]
		})
	}
}

func (h *Handler) parseGroupRequest(question dns.RR) groupRequest {
	var (
		request = groupRequest{
			name:   question.Header().Name,
			rrType: dns.RRToType(question),
		}

		subdomain = dnsutil.Trim(request.name, h.groupDomain)
		labels    = strings.Split(subdomain, ".")
	)

	request.group = labels[0]
	return request
}

func (h *Handler) serveGroup(response *dns.Msg, request groupRequest) {
	// we only communicate metadata via TXT records
	if request.rrType != dns.TypeTXT {
		response.Rcode = dns.RcodeRefused
		return
	}

	gps := h.locator.Groups()
	var rrs iter.Seq2[*service.Group, dns.RR]
	if len(request.group) > 0 {
		if g := gps.Get(request.group); g != nil {
			// only records for this group
			response.Answer = slices.Grow(response.Answer, g.LenRRs(request.rrType))
			rrs = g.RRs(request.rrType)
		}
	} else {
		// all records for all groups
		response.Answer = slices.Grow(response.Answer, gps.LenRRs(request.rrType))
		rrs = gps.RRs(request.rrType)
	}

	if rrs != nil {
		header := dns.Header{
			Name:  request.name,
			TTL:   h.jitterer.TTL(),
			Class: dns.ClassINET,
		}

		for _, rr := range rrs {
			*rr.Header() = header
			response.Answer = append(response.Answer, rr)
		}
	}
}

func (h *Handler) ServeDNS(ctx context.Context, writer dns.ResponseWriter, request *dns.Msg) {
	var (
		start  = time.Now()
		logger = h.logger.With(
			hashyzap.Request("request", request),
		)

		question = request.Question[0]
		response = request.Copy()
	)

	defer func() {
		logger.Info("request complete", zap.Duration("duration", time.Since(start)))
	}()

	logger.Info("request start")
	dnsutil.SetReply(response, request)
	response.Rcode = dns.RcodeSuccess // default
	defer h.writeResponse(logger, writer, response)

	if question.Header().Class != dns.ClassINET {
		logger.Error("unhandled class", zap.String("class", dnsutil.ClassToString(question.Header().Class)))
		response.Rcode = dns.RcodeRefused
		return
	}

	requestDomain := question.Header().Name
	switch {
	case dnsutil.IsBelow(h.endpointDomain, requestDomain):
		h.serveEndpoint(response, h.parseEndpointRequest(question))

	case dnsutil.IsBelow(h.groupDomain, requestDomain):
		h.serveGroup(response, h.parseGroupRequest(question))

	default:
		logger.Error("unrecognized domain", zap.String("domain", requestDomain))
		response.Rcode = dns.RcodeNotZone
	}
}
