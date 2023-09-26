package hashy

import (
	"errors"
	"net"
)

type Flusher interface {
	Flush() error
}

type ResponseWriter interface {
	Flusher
	AddResult(name []byte, groups, values []string) error
}

type responseWriter struct {
	remoteAddr net.Addr
	header     Header
	buffer     []byte
}

func (rw *responseWriter) Flush() (err error) {
	return
}

func (rw *responseWriter) AddResult(name []byte, groups, values []string) error {
	if len(groups) != len(values) {
		return errors.New("TODO: bad response")
	}

	// TODO: write a hash result

	return nil
}
