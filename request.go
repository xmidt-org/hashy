package hashy

import (
	"strings"

	"github.com/miekg/dns"
)

// Request represents a DNS request to hashy that hashy recognizes.
type Request struct {
	// ID is the value to hash. This is required.
	ID string

	// Nonce is an optional entropy value sent by the client.
	Nonce string

	// Group is the optional group name to restrict returned services.
	Group string

	// Qtype is the record type being requested.
	Qtype uint16

	// Origin is the domain that hashy recognized when it parsed this request.
	Origin string
}

// QuestionParser provides the logic to parse a DNS question into a request
// that hashy recognizes.
type QuestionParser struct {
	// Origin is the domain that is supported for hashy queries.
	Origin string
}

// Parse examines a DNS question and produces a Request. If the question
// is not supported, this method returns false.
func (qp *QuestionParser) Parse(q dns.Question) (r Request, supported bool) {
	if q.Qclass != dns.ClassINET && q.Qclass != dns.ClassANY {
		return
	}

	prefix, found := strings.CutSuffix(q.Name, qp.Origin)
	if !found {
		return
	}

	r.Qtype = q.Qtype
	r.Origin = qp.Origin
	r.ID, r.Group, _ = strings.Cut(prefix, ".")
	r.ID, r.Nonce, _ = strings.Cut(r.ID, "-")
	return
}
