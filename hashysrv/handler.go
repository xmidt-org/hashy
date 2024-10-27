package hashysrv

import (
	"github.com/miekg/dns"
	"github.com/xmidt-org/hashy/hashycfg"
	"github.com/xmidt-org/hashy/hashyzap"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.Logger
}

func (h *handler) ServeDNS(rw dns.ResponseWriter, request *dns.Msg) {
	h.logger.Info("request received", hashyzap.MsgField("request", request))

	var response dns.Msg
	response.SetReply(request)
	response.MsgHdr.Rcode = dns.RcodeRefused
	rw.WriteMsg(&response)
}

func NewHandler(l *zap.Logger, _ hashycfg.Config) dns.Handler {
	return &handler{
		logger: l,
	}
}
