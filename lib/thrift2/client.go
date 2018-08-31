package thrift2

import (
	"context"
	"errors"
	"sync"
)

var ErrPeerClosed = errors.New("thrift: peer closed")

type Client interface {
	Invoke(ctx context.Context, method string, arg, ret interface{}, options ...CallOption) error
	Close() error
}

type client struct {
	address string
	opts    options

	cpool Pool
	ppool sync.Pool
}

func NewClient(dialer Dialer, address string, opts ...Option) Client {
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
		if conn.IsReused() && err == ErrPeerClosed {
			// retry on reused & peer closed connection
			return cli.Invoke(ctx, method, arg, ret, options...)
		}
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

func (cli *client) Close() error {
	return cli.cpool.Close()
}
