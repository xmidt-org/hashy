package hashy

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
)

const (
	DefaultServerAddr                  = ":8080"
	DefaultServerNetwork               = "udp"
	DefaultServerMaxConcurrentRequests = 10
	DefaultServerMaxPacketSize         = 1500
)

const (
	serverStateNotStarted uint64 = iota
	serverStateRunning
	serverStateClosed
)

var (
	ErrServerNotStarted = errors.New("hashy: Server not started")
	ErrServerRunning    = errors.New("hashy: Server already running")
	ErrServerClosed     = errors.New("hashy: Server closed")
)

type Server struct {
	Addr                  string
	Network               string
	Handler               Handler
	MaxPacketSize         int
	MaxConcurrentRequests int

	state  atomic.Uint64
	ctx    context.Context
	cancel context.CancelFunc
	conn   net.PacketConn
	writer MessageWriter
}

func NewServer(cfg Config) (*Server, error) {
	return &Server{
		Addr:                  cfg.Address,
		Network:               cfg.Network,
		MaxPacketSize:         cfg.MaxPacketSize,
		MaxConcurrentRequests: cfg.MaxConcurrentRequests,
	}, nil
}

func (s *Server) listen() error {
	var (
		address = s.Addr
		network = s.Network
	)

	if len(address) == 0 {
		address = DefaultServerAddr
	}

	if len(network) == 0 {
		network = DefaultServerNetwork
	}

	conn, err := net.ListenPacket(network, address)
	if err != nil {
		return err
	}

	s.conn = conn
	s.writer = &syncMessageWriter{
		w: s.conn,
	}

	return nil
}

func (s *Server) newSemaphore() chan struct{} {
	depth := s.MaxConcurrentRequests
	if depth < 1 {
		depth = DefaultServerMaxConcurrentRequests
	}

	return make(chan struct{}, depth)
}

func (s *Server) newPacketBuffer() []byte {
	size := s.MaxPacketSize
	if size < 1 {
		size = DefaultServerMaxPacketSize
	}

	return make([]byte, size)
}

func (s *Server) serve() error {
	semaphore := s.newSemaphore()
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()

		case semaphore <- struct{}{}:
			// continue
		}

		packet := s.newPacketBuffer()
		n, remoteAddr, err := s.conn.ReadFrom(packet)
		if err != nil {
			s.Shutdown(context.Background())
			return err
		}

		go s.handle(semaphore, remoteAddr, packet[:n])
	}
}

func (s *Server) handle(semaphore <-chan struct{}, remoteAddr net.Addr, message []byte) {
	defer func() {
		<-semaphore
	}()

	var stream Stream[[]byte]
	header, payload, err := ReadHeaderBytes(message)
	if err == nil {
		stream, err = newStringStream(payload)
	}

	request := Request{
		RemoteAddr: remoteAddr,
		Header:     header,
		Names:      stream,
		ctx:        s.ctx,
	}

	rw := &responseWriter{
		remoteAddr: remoteAddr,
	}

	s.Handler.ServeHash(rw, &request)
}

func (s *Server) ListenAndServe() error {
	if !s.state.CompareAndSwap(serverStateNotStarted, serverStateRunning) {
		switch s.state.Load() {
		case serverStateRunning:
			return ErrServerRunning

		default:
			return ErrServerClosed
		}
	}

	err := s.listen()
	if err != nil {
		return err
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s.serve()
}

func (s *Server) Close() error {
	return s.Shutdown(context.Background())
}

func (s *Server) Shutdown(ctx context.Context) error {
	if !s.state.CompareAndSwap(serverStateRunning, serverStateClosed) {
		switch s.state.Load() {
		case serverStateNotStarted:
			return ErrServerNotStarted

		default:
			return ErrServerClosed
		}
	}

	s.cancel()
	return s.conn.Close()
}
