package thrift

import (
	"context"
	"errors"
	"sync"
)

var ErrPeerClosed = errors.New("thrift: peer closed")

type Invoker interface {
	Invoke(ctx context.Context, method string, arg, ret interface{}, options ...CallOption) error
	Close() error
}

type ProtocolInvoker interface {
	Invoker
	Protocol() *Protocol
}

type clientConnCtxKey struct{}
type clientProtocolCtxKey struct{}

// client implements the Invoker interface.
type client struct {
	address string
	opts    options

	cpool Pool
	ppool sync.Pool
}

func NewClient(dialer Dialer, address string, opts ...Option) *client {
	cli := &client{
		address: address,
		opts:    DefaultOptions,
	}
	for _, opt := range opts {
		cli.opts = opt(cli.opts)
	}
	cli.cpool = NewPool(dialer, opts...) // TODO
	cli.ppool.New = func() interface{} {
		return NewProtocol(nil, cli.opts)
	}
	return cli
}

func (cli *client) Invoke(ctx context.Context, method string, arg, ret interface{}, options ...CallOption) error {
	conn, err := cli.cpool.Take(ctx, cli.address)
	if err != nil {
		return err
	}
	defer cli.cpool.Put(conn)

	opts := cli.opts
	for _, opt := range options {
		opts = opt(opts)
	}
	_ = conn.SetReadTimeout(opts.rTimeout)  // shall not fail
	_ = conn.SetWriteTimeout(opts.wTimeout) // shall not fail
	if ctx.Done() != nil {
		var wg sync.WaitGroup
		ch := make(chan struct{})
		defer func() { close(ch); wg.Wait() }()
		wg.Add(1)
		go func() {
			select {
			case <-ctx.Done():
				conn.Close() // force cancel request
			case <-ch:
			}
			wg.Done()
		}()
	}

	prot := cli.ppool.Get().(*Protocol)
	prot.Reset(conn)
	defer cli.ppool.Put(prot)

	reqctx := context.WithValue(ctx, clientConnCtxKey{}, conn)
	reqctx = context.WithValue(reqctx, clientProtocolCtxKey{}, prot)
	err = invoke(reqctx, method, arg, ret)
	if err != nil && err == ErrPeerClosed && conn.IsReused() {
		// retry on reused & peer closed connection
		return cli.Invoke(ctx, method, arg, ret)
	}
	return err
}

func (cli *client) Close() error {
	return cli.cpool.Close()
}

func invoke(ctx context.Context, method string, arg, ret interface{}, options ...CallOption) (err error) {
	conn := ctx.Value(clientConnCtxKey{}).(Conn)
	prot := ctx.Value(clientProtocolCtxKey{}).(*Protocol)

	seqid := conn.NextSequence()
	_ = prot.WriteMessageBegin(method, CALL, seqid) // shall never fail
	if err = Write(arg, prot); err != nil {
		return err
	}
	if err = prot.Flush(); err != nil {
		return err
	}

	// Oneway method does not have result.
	if ret == nil {
		return nil
	}

	// Read the response.
	_, rt, rseq, err := prot.ReadMessageBegin()
	if err != nil {
		//if err == ErrPeerClosed && conn.IsReused() {
		//	// retry on reused & peer closed connection
		//	return cli.Invoke(ctx, method, arg, ret, options...)
		//}
		// TODO: for an EXCEPTION response, should not close the connection
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}
	if rseq != seqid {
		return ErrSeqMismatch
	}
	if rt == EXCEPTION {
		var exc ApplicationException
		if err = exc.Read(prot); err == nil {
			err = &exc
		}
		return err
	} else if rt != REPLY {
		return ErrMessageType
	}
	if err = Read(ret, prot); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}

	return nil
}

func NewProtocolInvokerFactory(dialer Dialer, opts ...Option) func(address string) (ProtocolInvoker, error) {
	factory := &factory{
		opts: DefaultOptions,
	}
	for _, opt := range opts {
		factory.opts = opt(factory.opts)
	}
	factory.cpool = NewPool(dialer, opts...) // TODO
	factory.ppool.New = func() interface{} {
		return NewProtocol(nil, factory.opts)
	}
	return factory.New
}

type factory struct {
	cpool Pool
	ppool sync.Pool
	opts  options
}

func (f *factory) New(address string) (ProtocolInvoker, error) {
	conn, err := f.cpool.Take(context.TODO(), address)
	if err != nil {
		return nil, err
	}
	// TODO: options
	prot := f.ppool.Get().(*Protocol)
	prot.Reset(conn)
	return &protocolInvoker{address: address, f: f, c: conn, p: prot}, nil
}

type protocolInvoker struct {
	address string
	f       *factory
	p       *Protocol
	c       Conn
}

func (c *protocolInvoker) Protocol() *Protocol { return c.p }

func (c *protocolInvoker) Close() error {
	c.f.ppool.Put(c.p)
	c.f.cpool.Put(c.c)
	return nil
}

func (c *protocolInvoker) Invoke(ctx context.Context, method string, arg, ret interface{}, options ...CallOption) (err error) {
	if conn, ok := ctx.Value(clientConnCtxKey{}).(Conn); !ok || conn != c.c {
		ctx = context.WithValue(ctx, clientConnCtxKey{}, c.c)
	}
	if prot, ok := ctx.Value(clientProtocolCtxKey{}).(*Protocol); !ok || prot != c.p {
		ctx = context.WithValue(ctx, clientProtocolCtxKey{}, c.p)
	}
	err = invoke(ctx, method, arg, ret)
	if err != nil && err == ErrPeerClosed && c.c.IsReused() {
		// retry on reused & peer closed connection
		c.Close()
		newInvoker, err := c.f.New(c.address)
		if err != nil {
			return err
		}

		*c = *(newInvoker.(*protocolInvoker))
		return c.Invoke(ctx, method, arg, ret)
	}
	return err
}
