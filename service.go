package hashy

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
)

// ServiceOption is a configurable option for a service.
type ServiceOption interface {
	apply(*Service) error
}

type serviceOptionFunc func(*Service) error

func (f serviceOptionFunc) apply(svc *Service) error { return f(svc) }

// WithIPs adds A or AAAA records to a service.
func WithIPs(ips ...net.IP) ServiceOption {
	return serviceOptionFunc(func(svc *Service) error {
		for _, ip := range ips {
			var rr dns.RR
			switch len(ip) {
			case net.IPv4len:
				rr = &dns.A{
					Hdr: dns.RR_Header{
						Name:   svc.name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    200, // TODO
					},
					A: ip,
				}

			case net.IPv6len:
				rr = &dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   svc.name,
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    200, // TODO
					},
					AAAA: ip,
				}

			default:
				return fmt.Errorf("Invalid IP: %s", ip)
			}

			svc.records = append(svc.records, rr)
		}

		return nil
	})
}

// Service is an endpoint that contains RR records to add to
// DNS messages. A Service is immutable once created.
type Service struct {
	name    string
	records []dns.RR
}

// Name returns the normalized name for this service.
func (s Service) Name() string {
	return s.name
}

func (s Service) createCNAME(request Request) *dns.CNAME {
	return &dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   request.Source.Name,
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    200, // TODO
		},
		Target: s.name,
	}
}

func (s Service) AddTo(response *dns.Msg, request Request) {
	qtype := request.Source.Qtype
	switch qtype {
	case dns.TypeANY:
		response.Answer = append(response.Answer, s.createCNAME(request))
		response.Answer = append(response.Answer, s.records...)

	case dns.TypeA:
		fallthrough

	case dns.TypeAAAA:
		fallthrough

	case dns.TypeCNAME:
		response.Answer = append(response.Answer, s.createCNAME(request))
		fallthrough

	default:
		for _, rr := range s.records {
			if rtype := rr.Header().Rrtype; qtype == rtype {
				response.Answer = append(response.Answer, rr)
			}
		}
	}
}

// NewService constructs a Service endpoint with the given name and option.
func NewService(name string, opts ...ServiceOption) (svc Service, err error) {
	svc.name = dns.CanonicalName(name)
	for _, o := range opts {
		if err = o.apply(&svc); err != nil {
			return
		}
	}

	return
}

// Services is an immutable set of services. Only (1) service of a given
// name may exist in a Services set.
type Services struct {
	m map[string]Service
}

// NewServices constructs an immutable Services set from the given list.
// Only (1) service with a given name will be present in the returned set.
// If any duplicates occur in the list, only the last service will be used.
func NewServices(list ...Service) (s Services) {
	s.m = make(map[string]Service, len(list))
	for _, svc := range list {
		s.m[svc.Name()] = svc
	}

	return
}

// Len returns the count of services in this set.
func (s Services) Len() int {
	return len(s.m)
}

// All provides iteration of the services in this set.
func (s Services) All(f func(Service) bool) {
	for _, service := range s.m {
		if !f(service) {
			return
		}
	}
}

// Merge returns a Services that is the union of a list of services with
// this Services set.
func (s Services) Merge(list ...Service) Services {
	switch {
	case len(list) == 0:
		return s

	case s.Len() == 0:
		return NewServices(list...)

	default:
		merged := Services{
			m: make(map[string]Service, s.Len()+len(list)),
		}

		for n, s := range s.m {
			merged.m[n] = s
		}

		for _, s := range list {
			merged.m[s.Name()] = s
		}

		return merged
	}
}

// Update produces a Services set that represents an update to this set.
// If this method returns true, the returned Services contains different
// services. If this method returns false, the list was not an update
// and this method returns this Services set.
func (s Services) Update(list ...Service) (Services, bool) {
	switch {
	case len(list) == 0:
		return s, false // no update

	case s.Len() == 0:
		// this Services is empty, so any non-empty list is an update
		return NewServices(list...), true

	default:
		updated := Services{
			m: make(map[string]Service, len(list)),
		}

		subset := true // whether updated is a subset of this Services
		for _, svc := range list {
			updated.m[svc.Name()] = svc
			if subset {
				_, subset = s.m[svc.Name()]
			}
		}

		if !subset || s.Len() != updated.Len() {
			// subset is false if there are names in the list that are not in updated
			// updated can be a subset and have fewer services
			return updated, true
		}

		return s, false
	}
}
