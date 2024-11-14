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

	// Origin is the domain that hashy recognized when it parsed this request.
	Origin string

	// Source is the DNS question that was parsed to produce this request.
	Source dns.Question
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
	r.Source = q
	if q.Qclass != dns.ClassINET && q.Qclass != dns.ClassANY {
		return
	}

	var prefix string
	if prefix, supported = strings.CutSuffix(q.Name, qp.Origin); supported {
		r.Origin = qp.Origin
		r.ID, r.Group, _ = strings.Cut(prefix, ".")
		r.ID, r.Nonce, _ = strings.Cut(r.ID, "-")
	}

	return
}
