// SPDX-FileCopyrightText: 2026 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashyzap

import (
	"codeberg.org/miekg/dns"
	"go.uber.org/zap"
)

func Request(fieldName string, request *dns.Msg) zap.Field {
	var questionField zap.Field
	if len(request.Question) > 0 {
		questionField = RR("question", request.Question[0])
	} else {
		questionField = zap.Skip()
	}

	return zap.Dict(
		fieldName,
		zap.Uint16("id", request.ID),
		zap.Uint16("udpSize", request.UDPSize),
		zap.Bool("recursionDesired", request.RecursionDesired),
		zap.Int("questionCount", len(request.Question)),
		questionField,
	)
}
