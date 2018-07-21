package thrift

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"

	"github.com/thrift-iterator/go"
)

// ErrServerClosed is returned by the Serve methods after a call to Stop
var ErrServerClosed = errors.New("thrift: Server closed")

type Processor interface {
	Process(ctx context.Context, reader *thrifter.Decoder, writer *thrifter.Encoder) error
}

type Server struct {
	processor Processor
	*serverOpts
}

func NewServer(p Processor, transport ServerTransport, options ...ServerOption) *Server {
	serverOptions := defaultServerOptions(transport)
	for _, option := range options {
		option(serverOptions)
	}
	return &Server{p, serverOptions}
}

// ServerTransport returns the Server transport
func (p *Server) ServerTransport() ServerTransport {
	return p.serverTransport
}

// InputTransportFactory returns the input transport factory
func (p *Server) InputTransportFactory() TransportFactory {
	return p.inputTransportFactory
}

// OutputTransportFactory returns the output transport factory
func (p *Server) OutputTransportFactory() TransportFactory {
	return p.outputTransportFactory
}

// Listen returns the Server transport listener
func (p *Server) Listen() error {
	return p.serverTransport.Listen()
}

// AcceptLoop runs the accept loop to handle requests
func (p *Server) AcceptLoop() error {
	for {
		client, err := p.serverTransport.Accept()
		if err != nil {
			select {
			case <-p.quit:
				return ErrServerClosed
			default:
			}
			// TODO: is this returning OK?
			return err
		}
		if client != nil {
			go p.processRequests(client)
		}
	}
}

// Serve starts serving requests
func (p *Server) Serve() error {
	err := p.Listen()
	if err != nil {
		return err
	}
	return p.AcceptLoop()
}

// ServeContext starts serving requests and uses a context to cancel
func (p *Server) ServeContext(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		p.Stop()
	}()
	err := p.Serve()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

// Stop stops the Server
func (p *Server) Stop() error {
	p.quit <- struct{}{}
	p.serverTransport.Interrupt()
	return nil
}

func (p *Server) processRequests(client Transport) error {
	var inputTransport, outputTransport Transport
	// TODO: Special case for HeaderTransport ?
	inputTransport = p.inputTransportFactory.GetTransport(client)
	outputTransport = p.outputTransportFactory.GetTransport(client)

	defer func() {
		if err := recover(); err != nil {
			p.log.Printf("panic in processor: %v: %s", err, debug.Stack())
		}
	}()
	if inputTransport != nil {
		defer inputTransport.Close()
	}
	if outputTransport != nil {
		defer outputTransport.Close()
	}

	reader := thrifter.NewDecoder(inputTransport, nil)
	writer := thrifter.NewEncoder(outputTransport)
	err := p.processor.Process(context.Background(), reader, writer)
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		// TODO: see fbthrift.Process(processor, inputProtocol, outputProtocol)
		p.log.Printf("processing failure: %s", err)
		return err
	}

	// Graceful exit. Client closed connection.
	return nil
}

type contextKey struct{ k string }

var ctxConnKey = contextKey{"ctxConnKey"}
