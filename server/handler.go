package server

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"codeberg.org/miekg/dns/rdata"
	"github.com/xmidt-org/hashy/service"
	"go.uber.org/zap"
)

const (
	// DefaultZoneDomain is the default DNS domain that hashy serves.
	DefaultZoneDomain = service.DefaultGeneratedEndpointDomain

	// DefaultZoneTTL is the default time-to-live for records generated within
	// hashy's zone.
	DefaultZoneTTL time.Duration = 5 * time.Minute

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

	// A indicates whether this request is asking for A records.
	A bool

	// AAAA indicates whether this request is asking for AAAA records.
	AAAA bool
}

// Handler is the main DNS handler for hashy.
//
// This type does not implement dns.Handler. ServeDNS takes a logger that is enriched by Middleware.
type Handler struct {
	Domain  string
	Locator *service.Locator
	TTL     uint32
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
	parentLabels := dnsutil.Labels(h.Domain)
	requestLabels := dnsutil.Labels(requestDomain)
	if dnsutil.Common(h.Domain, requestDomain) != parentLabels {
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

	switch question.Data().(type) {
	case rdata.A:
		hr.A = true

	case rdata.AAAA:
		hr.AAAA = true
	}

	// NOTE: an empty Name will cause response.Pack() to panic
	hr.Name = requestDomain

	return
}

// handleRequest handles a valid HashRequest.
func (h *Handler) handleRequest(_ context.Context, logger *zap.Logger, response *dns.Msg, request HashRequest) {
	logger.Debug("handleRequest", zap.Any("request", request))
	response.Rcode = dns.RcodeSuccess

	for _, endpoint := range h.Locator.FindString(request.Object) {
		if request.A {
			for addr := range endpoint.A() {
				response.Answer = append(response.Answer, &dns.A{
					Hdr: dns.Header{
						Name:  request.Name,
						TTL:   h.TTL, // TODO: jitter?
						Class: dns.ClassINET,
					},
					A: rdata.A{
						Addr: addr,
					},
				})
			}
		}
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

	h.handleRequest(ctx, logger, response, hashRequest)
	if err := response.Pack(); err != nil {
		logger.Error("unable to pack response", zap.Error(err))
		return
	}

	io.Copy(writer, response)
}
