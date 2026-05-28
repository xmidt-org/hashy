package server

import (
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy/service"
	"go.uber.org/zap"
)

const (
	// DefaultZoneDomain is the default DNS domain that hashy serves.
	DefaultZoneDomain = service.DefaultGeneratedEndpointDomain

	// DefaultZoneTTL is the default time-to-live for records generated within
	// hashy's zone.
	DefaultZoneTTL time.Duration = 5 * time.Minute

	// GroupsLabel defines the subdomain of the zone domain that handles group DNS lookups.
	GroupsLabel = "groups"

	// HashLabelPrefix is the prefix, without a hyphen, of domain names requesting
	// a hash. This prefix can be used to make hash objects that start with numbers
	// into a valid hostname.
	HashLabelPrefix = "id"
)

var (
	errQuestionCount = errors.New("only one (1) question is supported")
	errDomain        = errors.New("hashy does not handle that domain")
	errClass         = errors.New("hashy does not handle that class")
	errLabels        = errors.New("exadtly one (1) hash label is required")
)

type HashRequest struct {
	Name string

	Prefix string
	Object string
	Extra  []string

	Type uint16
}

// Handler is the main DNS handler for hashy.
//
// This type does not implement dns.Handler. ServeDNS takes a logger that is enriched by Middleware.
type Handler struct {
	zoneDomain   string
	groupsDomain string
	locator      *service.Locator
	ttl          uint32
}

// parseRequest validates and parses a request dns.Message into a HashRequest.
func (h *Handler) parseRequest(request *dns.Msg) (hr HashRequest, err error) {
	request.Unpack()
	if len(request.Question) != 1 {
		err = errQuestionCount
		return
	}

	question := request.Question[0]
	if question.Header().Class != dns.ClassINET {
		err = errClass
		return
	}

	// the question domain must have exactly (1) more label, which is the hash request string,
	// and have the subdomain.
	requestDomain := question.Header().Name
	parentLabels := dnsutil.Labels(h.zoneDomain)
	requestLabels := dnsutil.Labels(requestDomain)
	if dnsutil.Common(h.zoneDomain, requestDomain) != parentLabels {
		// the requestDomain is not a proper subdomain of what this handler serves
		err = errDomain
		return
	} else if requestLabels != (parentLabels + 1) {
		// the request wasn't of the form {hash label}.{domain}
		err = errLabels
		return
	}

	firstLabel, _, _ := strings.Cut(question.Header().Name, ".")
	parts := strings.Split(firstLabel, "-")
	switch {
	case len(parts) == 1:
		hr = HashRequest{
			Object: parts[0],
		}

	case parts[0] == HashLabelPrefix:
		hr = HashRequest{
			Prefix: parts[0],
			Object: parts[1],
			Extra:  parts[2:],
		}

	default:
		hr = HashRequest{
			Object: parts[0],
			Extra:  parts[1:],
		}
	}

	// NOTE: an empty Name will cause response.Pack() to panic
	hr.Name = requestDomain
	hr.Type = dns.RRToType(question)
	return
}

// handleRequest handles a valid HashRequest.
func (h *Handler) handleRequest(_ context.Context, response *dns.Msg, request HashRequest) {
	response.Rcode = dns.RcodeSuccess
	endpoints := h.locator.FindString(request.Object)
	response.Answer = slices.Grow(response.Answer, endpoints.LenRRs(request.Type))

	header := dns.Header{
		Name:  request.Name,
		TTL:   h.ttl, // TODO: jitter?
		Class: dns.ClassINET,
	}

	for _, rr := range endpoints.RRs(request.Type) {
		*rr.Header() = header
		response.Answer = append(response.Answer, rr)
	}
}

// ServeDNS contains hashy's main logic. It responds to DNS queries by using a service.Locator to generate
// DNS responses.
func (h *Handler) ServeDNS(ctx context.Context, logger *zap.Logger, writer dns.ResponseWriter, request *dns.Msg) {
	response := request.Copy()
	dnsutil.SetReply(response, request)
	hashRequest, err := h.parseRequest(request)

	if err != nil {
		if errors.Is(err, errClass) {
			response.Rcode = dns.RcodeRefused
		} else {
			response.Rcode = dns.RcodeNameError
		}

		logger.Error("bad request", zap.Error(err))
		return
	}

	h.handleRequest(ctx, response, hashRequest)
	if err := response.Pack(); err != nil {
		logger.Error("unable to pack response", zap.Error(err))
		return
	}

	io.Copy(writer, response)
}
