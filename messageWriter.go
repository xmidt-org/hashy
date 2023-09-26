package hashy

import (
	"net"
	"sync"
)

// MessageWriter defines the behavior of a message-oriented writer.
type MessageWriter interface {
	// WriteTo has the same contract as net.PacketConn.WriteTo.
	WriteTo([]byte, net.Addr) (int, error)
}

type syncMessageWriter struct {
	lock sync.Mutex
	w    MessageWriter
}

func (sw *syncMessageWriter) WriteTo(b []byte, a net.Addr) (int, error) {
	defer sw.lock.Unlock()
	sw.lock.Lock()
	return sw.w.WriteTo(b, a)
}

func NewSyncMessageWriter(w MessageWriter) MessageWriter {
	if _, ok := w.(*syncMessageWriter); ok {
		return w
	}

	return &syncMessageWriter{
		w: w,
	}
}
