package hashysrv

import (
	"net"

	"github.com/miekg/dns"
	"github.com/xmidt-org/hashy"
	"github.com/xmidt-org/hashy/hashycfg"
	"github.com/xmidt-org/hashy/hashyzap"
	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	parser  hashy.QuestionParser
	service hashy.Service
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
	if hRequest, applies := h.parser.Parse(request.Question[0]); applies {
		h.logger.Info("processing request", zap.String("id", hRequest.ID))
		response.SetReply(request)
		h.service.AddTo(&response, hRequest)
	} else {
		response.SetRcode(request, dns.RcodeRefused)
	}

	rw.WriteMsg(&response)
}

func NewHandler(l *zap.Logger, _ hashycfg.Config) (h *Handler, err error) {
	h = &Handler{
		logger: l,
		parser: hashy.QuestionParser{
			Origin: ".device.",
		},
	}

	h.service, err = hashy.NewService(
		"talaria-123.xmidt.comcast.net",
		hashy.WithIPs(
			net.IP{127, 0, 0, 1},
			net.IP{0xfc, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		),
	)

	return
}
