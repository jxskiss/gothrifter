package thrift

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

// ErrServerClosed is returned by the Serve methods after a call to Stop
var ErrServerClosed = errors.New("thrift: server closed")

type Processor interface {
	Process(ctx context.Context, r Reader, w Writer) error
}

type Server struct {
	processor Processor
	opts      options

	listener net.Listener
	ppool    sync.Pool
	n        int64
	quit     chan struct{}
}

func NewServer(p Processor, options ...Option) *Server {
	s := &Server{
		processor: p,
		opts:      DefaultOptions,
		quit:      make(chan struct{}, 1),
	}
	s.opts.maxActive = 50000
	for _, option := range options {
		s.opts = option(s.opts)
	}
	s.ppool.New = func() interface{} {
		return NewProtocol(nil, s.opts)
	}
	return s
}

// Listen returns the Server transport listener
func (p *Server) Listen(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	p.listener = listener
	return nil
}

// Serve runs the accept loop to handle requests.
func (p *Server) Serve() error {
	for {
		client, err := p.listener.Accept()
		if err != nil {
			log.Println("accept error:", err) // TODO
			continue
		}
		select { // TODO
		case <-p.quit:
			return ErrServerClosed
		default:
			// pass
		}
		cur := atomic.AddInt64(&p.n, 1)
		if p.opts.maxActive > 0 && cur > int64(p.opts.maxActive) {
			log.Printf("server: max active connection execcded %d/%d\n", cur, p.opts.maxActive)
			atomic.AddInt64(&p.n, -1)
			client.Close()
			continue
		}
		go p.process(client) // TODO: what about errors?
	}
}

func (p *Server) ListenAndServe(addr string) error {
	if err := p.Listen(addr); err != nil {
		return err
	}
	return p.Serve()
}

// Stop stops the Server.
func (p *Server) Stop() error {
	p.quit <- struct{}{}
	return nil
}

func (p *Server) process(client net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 64<<10)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("server: panic serving %s: %v\n%s\n", client.RemoteAddr(), err, buf)
		}
		client.Close() // potential errors ignored
		atomic.AddInt64(&p.n, -1)
	}()
	prot := p.ppool.Get().(*Protocol)
	defer p.ppool.Put(prot)

	prot.Reset(client)
	ctx := context.Background()
	ctx = context.WithValue(ctx, protocolCtxKey{}, prot)
	ctx = context.WithValue(ctx, remoteAddrCtxKey{}, client.RemoteAddr().String())
	if err := p.processor.Process(ctx, prot, prot); err != nil {
		if err != io.EOF && !isForciblyClosed(err) {
			log.Printf("server: process client %s error: %s\n", client.RemoteAddr(), err)
		}
	}
}

type (
	protocolCtxKey   struct{}
	remoteAddrCtxKey struct{}
)

func ProtocolFromCtx(ctx context.Context) *Protocol {
	if p, ok := ctx.Value(protocolCtxKey{}).(*Protocol); ok {
		return p
	}
	return nil
}

func RemoteAddrFromCtx(ctx context.Context) string {
	if addr, ok := ctx.Value(remoteAddrCtxKey{}).(string); ok {
		return addr
	}
	return ""
}

func isForciblyClosed(err error) bool {
	if e, ok := err.(*net.OpError); ok {
		return strings.Contains(e.Err.Error(), "forcibly closed")
	}
	return false
}
