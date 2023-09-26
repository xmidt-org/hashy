package hashy

import (
	"context"
	"net"
)

type Request struct {
	RemoteAddr net.Addr
	Header     Header
	Names      Stream[[]byte]

	ctx context.Context
}

func (r *Request) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}

	return r.ctx
}
