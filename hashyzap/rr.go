// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashyzap

import (
	"codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RRObjectMarshaler struct {
	rr dns.RR
}

func (om RRObjectMarshaler) MarshalLogObject(enc zapcore.ObjectEncoder) (err error) {
	header := om.rr.Header()
	enc.AddString(
		"name", header.Name,
	)

	enc.AddString(
		"type", dnsutil.TypeToString(dns.RRToType(om.rr)),
	)

	enc.AddString(
		"class", dnsutil.ClassToString(header.Class),
	)

	return
}

// RR produces a zap field for a DNS RR
func RR(fieldName string, rr dns.RR) zap.Field {
	return zap.Object(
		fieldName,
		RRObjectMarshaler{rr: rr},
	)
}
