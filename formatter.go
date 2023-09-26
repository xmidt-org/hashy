package hashy

import (
	"encoding/binary"
	"io"
)

// formatter is a low-level writer for hashy data.  Byte sequences, including
// strings, are appended with varint length indicators.
type formatter struct {
	dst io.Writer

	sw io.StringWriter
	bw io.ByteWriter

	// buffer is long enough to hold a maximal uint64 varint
	buffer [10]byte
}

func (f *formatter) write(b []byte) (n int, err error) {
	n, err = f.dst.Write(
		binary.AppendUvarint(f.buffer[:], uint64(len(b))),
	)

	if err == nil {
		var written int
		written, err = f.dst.Write(b)
		n += written
	}

	return
}

func (f *formatter) writeByte(c byte) error {
	if f.bw != nil {
		return f.bw.WriteByte(c)
	}

	f.buffer[0] = c
	_, err := f.dst.Write(f.buffer[:1])
	return err
}

func (f *formatter) writeString(v string) (n int, err error) {
	n, err = f.dst.Write(
		binary.AppendUvarint(f.buffer[:], uint64(len(v))),
	)

	var written int
	if f.sw != nil {
		written, err = f.sw.WriteString(v)
	} else {
		written, err = f.dst.Write([]byte(v))
	}

	n += written
	return
}

func newFormatter(w io.Writer) *formatter {
	f := &formatter{
		dst: w,
	}

	f.sw, _ = w.(io.StringWriter)
	f.bw, _ = w.(io.ByteWriter)
	return f
}
