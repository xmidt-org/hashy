package hashy

import (
	"net"
	"sync"
)

type Flusher interface {
	Flush() error
}

// Writer defines the behavior of a packet-oriented writer.
type Writer interface {
	// WriteTo has the same contract as net.PacketConn.WriteTo.
	WriteTo([]byte, net.Addr) (int, error)
}

type syncWriter struct {
	lock sync.Mutex
	w    Writer
}

func (sw *syncWriter) WriteTo(b []byte, a net.Addr) (int, error) {
	defer sw.lock.Unlock()
	sw.lock.Lock()
	return sw.w.WriteTo(b, a)
}

type ResponseWriter interface {
	Flusher
	Add(DeviceName, Datacenter, string)
}

type responseWriter struct {
	remoteAddr net.Addr
	writer     Writer
	hashes     DeviceHashes
}

func (rw *responseWriter) Add(name DeviceName, dc Datacenter, value string) {
	rw.hashes.Add(name, dc, value)
}

func (rw *responseWriter) Flush() (err error) {
	var message []byte
	message, err = MarshalBytes(rw.hashes)
	if err == nil {
		_, err = rw.writer.WriteTo(message, rw.remoteAddr)
	}

	return
}
