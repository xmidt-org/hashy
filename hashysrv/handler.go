// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package hashysrv

import (
	"github.com/miekg/dns"
	"github.com/xmidt-org/hashy/hashycfg"
	"github.com/xmidt-org/hashy/hashyzap"
	"go.uber.org/zap"
)

type Handler struct {
	logger *zap.Logger
}

func (h *Handler) WithLogFields(f ...zap.Field) *Handler {
	with := new(Handler)
	*with = *h
	with.logger = with.logger.With(f...)
	return with
}

func (h *Handler) ServeDNS(rw dns.ResponseWriter, request *dns.Msg) {
	h.logger.Info("request received", hashyzap.MsgField("request", request))

	var response dns.Msg
	response.SetReply(request)
	response.MsgHdr.Rcode = dns.RcodeRefused
	rw.WriteMsg(&response)
}

func NewHandler(l *zap.Logger, _ hashycfg.Config) *Handler {
	return &Handler{
		logger: l,
	}
}
