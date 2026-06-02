package server

import (
	"context"
	"errors"
	"slices"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/service"
	"go.uber.org/zap"
)

// EndpointRequest holds the information from a DNS question for responding with located endpoints.
type EndpointRequest struct {
	name   string
	groups []string

	prefix string
	object string
	extra  []string

	rrType uint16
}

// ParseEndpointRequest parses the question into an endpoint object carrying the
// information to satisfy and endpoint hash.
func ParseEndpointRequest(question dns.RR, domain string) EndpointRequest {
	var (
		request = EndpointRequest{
			name:   question.Header().Name,
			rrType: dns.RRToType(question),
		}

		subdomain = dnsutil.Trim(request.name, domain)
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

type EndpointHandlerOption interface {
	applyToEndpointHandler(*EndpointHandler) error
}

type endpointHandlerOptionFunc func(*EndpointHandler) error

func (f endpointHandlerOptionFunc) applyToEndpointHandler(eh *EndpointHandler) error { return f(eh) }

func WithEndpointLocator(locator *service.Locator) EndpointHandlerOption {
	return endpointHandlerOptionFunc(func(eh *EndpointHandler) error {
		eh.locator = locator
		return nil
	})
}

func WithEndpointJitterer(j *hashy.TTLJitterer) EndpointHandlerOption {
	return endpointHandlerOptionFunc(func(eh *EndpointHandler) error {
		eh.jitterer = j
		return nil
	})
}

// EndpointHandler produces address records and other metadata based on a consistent hash.
type EndpointHandler struct {
	locator  *service.Locator
	jitterer *hashy.TTLJitterer
}

func NewEndpointHandler(opts ...EndpointHandlerOption) (*EndpointHandler, error) {
	eh := new(EndpointHandler)
	for _, o := range opts {
		if err := o.applyToEndpointHandler(eh); err != nil {
			return nil, err
		}
	}

	if eh.locator == nil {
		return nil, errors.New("a locator is required")
	}

	if eh.jitterer == nil {
		// we know that this call is valid and won't return errors
		eh.jitterer, _ = hashy.NewTTLJitterer(hashy.DurationToSeconds(DefaultZoneTTL), 0.0)
	}

	return eh, nil
}

func (eh *EndpointHandler) ServeRequest(_ context.Context, _ *zap.Logger, response *dns.Msg, request EndpointRequest) {
	endpoints := eh.locator.FindString(request.object, request.groups...)
	response.Answer = slices.Grow(response.Answer, endpoints.LenRRs(request.rrType))

	header := dns.Header{
		Name:  request.name,
		TTL:   eh.jitterer.TTL(),
		Class: dns.ClassINET,
	}

	for _, rr := range endpoints.RRs(request.rrType) {
		*rr.Header() = header
		response.Answer = append(response.Answer, rr)
	}

	hashy.Shuffle(response.Answer)
}
