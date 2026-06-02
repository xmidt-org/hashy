// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"errors"
	"iter"
	"slices"
	"strings"

	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/service"
	"go.uber.org/zap"
)

// GroupRequest holds the information from a DNS question for responding with group metadata.
type GroupRequest struct {
	name string

	group  string
	rrType uint16
}

func ParseGroupRequest(question dns.RR, domain string) GroupRequest {
	var (
		request = GroupRequest{
			name:   question.Header().Name,
			rrType: dns.RRToType(question),
		}

		subdomain = dnsutil.Trim(request.name, domain)
		labels    = strings.Split(subdomain, ".")
	)

	request.group = labels[0]
	return request
}

type GroupHandlerOption interface {
	applyToGroupHandler(*GroupHandler) error
}

type groupHandlerOptionFunc func(*GroupHandler) error

func (f groupHandlerOptionFunc) applyToGroupHandler(gh *GroupHandler) error { return f(gh) }

func WithGroupLocator(locator *service.Locator) GroupHandlerOption {
	return groupHandlerOptionFunc(func(gh *GroupHandler) error {
		gh.locator = locator
		return nil
	})
}

func WithGroupJitterer(j *hashy.TTLJitterer) GroupHandlerOption {
	return groupHandlerOptionFunc(func(gh *GroupHandler) error {
		gh.jitterer = j
		return nil
	})
}

type GroupHandler struct {
	locator  *service.Locator
	jitterer *hashy.TTLJitterer
}

func NewGroupHandler(opts ...GroupHandlerOption) (*GroupHandler, error) {
	gh := new(GroupHandler)
	for _, o := range opts {
		if err := o.applyToGroupHandler(gh); err != nil {
			return nil, err
		}
	}

	if gh.locator == nil {
		return nil, errors.New("a locator is required")
	}

	if gh.jitterer == nil {
		// we know that this call is valid and won't return errors
		gh.jitterer, _ = hashy.NewTTLJitterer(hashy.DurationToSeconds(DefaultZoneTTL), 0.0)
	}

	return gh, nil
}

func (gh *GroupHandler) ServeRequest(_ context.Context, _ *zap.Logger, response *dns.Msg, request GroupRequest) {
	// we only communicate metadata via TXT records
	if request.rrType != dns.TypeTXT {
		response.Rcode = dns.RcodeRefused
		return
	}

	gps := gh.locator.Groups()
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
			TTL:   gh.jitterer.TTL(),
			Class: dns.ClassINET,
		}

		for _, rr := range rrs {
			*rr.Header() = header
			response.Answer = append(response.Answer, rr)
		}
	}
}
