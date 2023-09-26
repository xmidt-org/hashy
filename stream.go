package hashy

import (
	"encoding/binary"
	"errors"
	"io"
)

type Stream[T any] interface {
	Next() (T, error)
}

type stringStream struct {
	count  int
	buffer []byte
}

func (ss *stringStream) Next() ([]byte, error) {
	if len(ss.buffer) == 0 {
		return nil, io.EOF
	}

	l, n := binary.Uvarint(ss.buffer)
	length := int(l)
	if n <= 0 || len(ss.buffer) < n+length {
		return nil, io.ErrUnexpectedEOF
	}

	b := ss.buffer[n : n+length]
	ss.buffer = ss.buffer[n+length:]
	return b, nil
}

func newStringStream(b []byte) (Stream[[]byte], error) {
	n, count := binary.Uvarint(b)
	if n <= 0 {
		return nil, errors.New("TODO: badly formatted query")
	}

	return &stringStream{
		count:  int(count),
		buffer: b[n:],
	}, nil
}
