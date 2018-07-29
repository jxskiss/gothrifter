package thrift

import (
	"context"
	"io"
	"sync"

	"github.com/thrift-iterator/go/protocol"
)

type Client interface {
	Invoke(ctx context.Context, method string, arg, ret interface{}, options ...CallOption) error
	Close() error
}

type client struct {
	pool    Pool
	address string

	opts clientOpts
}

func NewClient(dialer Dialer, address string, options ...ClientOption) Client {
	pool := NewPool(dialer, options...)
	cli := &client{
		pool:    pool,
		address: address,
		opts:    defaultClientOptions(),
	}
	for _, option := range options {
		cli.opts = option(cli.opts)
	}
	return cli
}

func (cli *client) Invoke(ctx context.Context, method string, arg, ret interface{}, options ...CallOption) (err error) {
	conn, err := cli.pool.Take(ctx, cli.address)
	if err != nil {
		return err
	}
	defer cli.pool.Put(conn)

	opts := defaultCallOptions(cli.opts)
	for _, option := range options {
		opts = option(opts)
	}
	if err = conn.SetReadTimeout(opts.rTimeout); err != nil {
		return err
	}
	if err = conn.SetWriteTimeout(opts.wTimeout); err != nil {
		return err
	}

	// TODO: SSL
	var transport io.ReadWriter = conn
	if cli.opts.tFactory != nil {
		socket := NewSocketFromConnTimeout(conn, 0)
		transport = cli.opts.tFactory.GetTransport(socket)
	}

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

	seqId := protocol.SeqId(conn.NextSequence())

	// Write request.
	reqHeader := protocol.MessageHeader{
		MessageName: method,
		MessageType: protocol.MessageTypeCall,
		SeqId:       seqId,
	}
	encoder := cli.opts.tCfg.NewEncoder(transport)
	if err = encoder.EncodeMessageHeader(reqHeader); err != nil {
		if conn.IsReused() && err == errPeerClosed {
			// retry on reused & peer closed connection
			return cli.Invoke(ctx, method, arg, ret)
		}
		if errCtx := ctx.Err(); errCtx != nil {
			return errCtx
		}
		return err
	}
	if err = encoder.Encode(arg); err != nil {
		return err
	}

	// Oneway method without result.
	if ret == nil {
		return nil
	}

	// Read response.
	decoder := cli.opts.tCfg.NewDecoder(transport, nil)
	rspHeader, err := decoder.DecodeMessageHeader()
	if err != nil {
		if conn.IsReused() && err == errPeerClosed {
			// retry on reused & peer closed connection
			return cli.Invoke(ctx, method, arg, ret)
		}
		if errCtx := ctx.Err(); errCtx != nil {
			return errCtx
		}
		return err
	}
	if rspHeader.SeqId != seqId {
		return ErrSeqMismatch
	}
	if t := rspHeader.MessageType; t == protocol.MessageTypeException {
		var exc ApplicationException
		if err = decoder.Decode(&exc); err == nil {
			err = &exc
		}
		return err
	} else if t != protocol.MessageTypeReply {
		return ErrMessageType
	}
	if err = decoder.Decode(ret); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}

	return nil
}

func (cli *client) Close() error {
	return cli.pool.Close()
}
