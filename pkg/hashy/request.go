package hashy

import (
	"context"
	"net"
)

type Request struct {
	RemoteAddr  net.Addr
	DeviceNames DeviceNames

	ctx context.Context
}

func (r *Request) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}

	return r.ctx
}

func NewRequest(names DeviceNames) *Request {
	return &Request{
		DeviceNames: names,
		ctx:         context.Background(),
	}
}
