package hashy

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

var (
	ErrTruncation           = errors.New("hashy: truncated message")
	ErrStringTooLarge       = fmt.Errorf("hashy: string longer than %d bytes", math.MaxUint8)
	ErrSliceTooLarge        = fmt.Errorf("hashy: cannot have more than %d elements in a slice", math.MaxUint8)
	ErrPacketLengthExceeded = errors.New("hashy: buffer length exceeded")
)

const (
	ResponseFlag uint8 = 0b10000000
	NoChangeFlag uint8 = 0b01000000
	TruncateFlag uint8 = 0b00000001

	// HeaderSize is the required bytes for the Header
	HeaderSize int = 12
)

func appendString[T ~[]byte | ~string](b []byte, v T) []byte {
	b = binary.AppendUvarint(b, uint64(len(v)))
	b = append(b, v...)
	return b
}

func appendStrings[T ~[]byte | ~string](b []byte, s []T) []byte {
	b = binary.AppendUvarint(b, uint64(len(s)))
	for _, v := range s {
		b = appendString(b, v)
	}

	return b
}

func appendSlice[T any](b []byte, f func([]byte, T) []byte, s []T) []byte {
	b = binary.AppendUvarint(b, uint64(len(s)))
	for _, v := range s {
		b = f(b, v)
	}

	return b
}

type Header struct {
	Version  byte
	Id       uint16
	Response bool
	Truncate bool
	NoChange bool
	Tag      uint64
}

func (h Header) Flags() uint8 {
	var f uint8
	if h.Response {
		f |= ResponseFlag
	}

	if h.Truncate {
		f |= TruncateFlag
	}

	if h.NoChange {
		f |= NoChangeFlag
	}

	return f
}

func (h *Header) SetFlags(f uint8) {
	h.Response = (f&ResponseFlag != 0)
	h.Truncate = (f&TruncateFlag != 0)
	h.NoChange = (f&NoChangeFlag != 0)
}

func AppendHeader(b []byte, h Header) []byte {
	b = append(b, h.Version)
	b = binary.BigEndian.AppendUint16(b, h.Id)
	b = append(b, h.Flags())
	b = binary.BigEndian.AppendUint64(b, h.Tag)

	return b
}

func ReadHeader(r io.Reader) (h Header, n int, err error) {
	var b [HeaderSize]byte
	if n, err = r.Read(b[:]); n < HeaderSize && (err == nil || err == io.EOF) {
		err = io.ErrUnexpectedEOF
	}

	if err == nil {
		h, _, err = ReadHeaderBytes(b[:])
	}

	return
}

func ReadHeaderBytes(b []byte) (h Header, remaining []byte, err error) {
	remaining = b[HeaderSize:]
	h.Version = b[0]
	h.Id = binary.BigEndian.Uint16(b[1:3])
	h.SetFlags(b[3])
	h.Tag = binary.BigEndian.Uint64(b[4:12])
	return
}

type HashQuery struct {
	Header Header
	Names  []string
}

func AppendHashQuery(b []byte, hq HashQuery) []byte {
	b = AppendHeader(b, hq.Header)
	b = appendStrings(b, hq.Names)
	return b
}

// HashValue is the tuple of a group and the value that resulted
// from hashing a name.
type HashValue struct {
	Group string
	Value string
}

func AppendHashValue(b []byte, hv HashValue) []byte {
	b = appendString(b, hv.Group)
	b = appendString(b, hv.Value)
	return b
}

// HashResult is the result of applying hashes for a value to each
// known group.
type HashResult struct {
	Name   string
	Values []HashValue
}

func AppendHashResult(b []byte, hr HashResult) []byte {
	b = appendString(b, hr.Name)
	b = appendSlice(
		b,
		AppendHashValue,
		hr.Values,
	)

	return b
}

type HashAnswer struct {
	Header  Header
	Results []HashResult
}

func AppendHashAnswer(b []byte, ha HashAnswer) []byte {
	b = AppendHeader(b, ha.Header)
	b = appendSlice(
		b,
		AppendHashResult,
		ha.Results,
	)

	return b
}
