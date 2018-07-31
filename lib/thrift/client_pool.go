package thrift

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Dialer imitates net.Dial. Dialer is assumed to yield connections that are
// safe for use by multiple concurrent goroutines.
type Dialer func(ctx context.Context, address string) (net.Conn, error)

func (d Dialer) WithRetry(n int) Dialer {
	return func(ctx context.Context, address string) (conn net.Conn, err error) {
		for i := n; ; i++ {
			conn, err = d(ctx, address)
			if err == nil || ctx.Err() != nil || i <= 0 {
				return
			}
		}
	}
}

var StdDialer Dialer = func(ctx context.Context, address string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, "tcp", address)
}

func NewTimeoutDialer(timeout time.Duration) Dialer {
	return func(ctx context.Context, address string) (net.Conn, error) {
		var d = net.Dialer{Timeout: timeout}
		return d.DialContext(ctx, "tcp", address)
	}
}

func NewTlsDialer(tlsCfg *tls.Config) Dialer {
	return func(ctx context.Context, address string) (net.Conn, error) {
		return tls.Dial("tcp", address, tlsCfg)
	}
}

type Pool interface {
	Take(ctx context.Context, address string) (Conn, error)
	Put(conn Conn) error
	Close() error
}

type Conn interface {
	net.Conn
	SetTimeout(t time.Duration) error
	SetReadTimeout(t time.Duration) error
	SetWriteTimeout(t time.Duration) error

	NextSequence() int32
	IsReused() bool
}

type pool struct {
	dial Dialer
	opts clientOpts

	mu       sync.Mutex
	peers    sync.Map
	isClosed bool
}

func NewPool(dialer Dialer, opts ...ClientOption) Pool {
	pool := &pool{dial: dialer}
	pool.opts = defaultClientOptions()
	for _, opt := range opts {
		pool.opts = opt(pool.opts)
	}
	return pool
}

func (p *pool) Take(ctx context.Context, address string) (Conn, error) {
	var pp *peer
	if x, ok := p.peers.Load(address); ok {
		pp = x.(*peer)
		return pp.take(ctx)
	}

	p.mu.Lock()
	pp = &peer{
		pool:    p,
		address: address,
	}
	if p.opts.maxIdle > 0 && !p.isClosed {
		pp.free = make(chan *clientconn, p.opts.maxIdle)
	}
	if x, loaded := p.peers.LoadOrStore(address, pp); loaded {
		pp = x.(*peer)
	}
	p.mu.Unlock()
	return pp.take(ctx)
}

func (p *pool) Put(conn Conn) error {
	cc := conn.(*clientconn)
	return cc.peer.put(cc)
}

func (p *pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClosed {
		return nil
	}
	p.isClosed = true
	p.peers.Range(func(k, v interface{}) bool {
		pp := v.(*peer)
		pp.close()
		return true
	})
	return nil
}

type peer struct {
	pool    *pool
	address string

	mu     sync.RWMutex
	free   chan *clientconn
	active int32 // atomic
}

func (p *peer) take(ctx context.Context) (*clientconn, error) {
	select {
	case conn := <-p.free:
		if conn.isExpired() {
			atomic.AddInt32(&p.active, -1)
			conn.Conn.Close()
			return p.take(ctx)
		}
		return conn, nil
	default:
		// pass
	}
	active := atomic.AddInt32(&p.active, 1)
	if p.pool.opts.maxActive > 0 && int(active) > p.pool.opts.maxActive {
		atomic.AddInt32(&p.active, -1)
		return nil, errTooManyConn
	}

	// Establish new connection.
	netConn, err := p.pool.dial(ctx, p.address)
	if err != nil {
		atomic.AddInt32(&p.active, -1)
		return nil, err
	}
	conn := newClientconn(netConn, p)
	return conn, err
}

func (p *peer) put(conn *clientconn) error {
	if err := conn.getError(); err != nil {
		atomic.AddInt32(&p.active, -1)
		return conn.Conn.Close()
	}
	conn.usedAt = time.Now()
	p.mu.RLock()
	defer p.mu.RUnlock()
	select {
	case p.free <- conn:
		return nil
	default:
		atomic.AddInt32(&p.active, -1)
		return conn.Conn.Close()
	}
}

func (p *peer) close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.free == nil {
		return nil
	}
	for {
		select {
		case conn := <-p.free:
			if err := conn.Conn.Close(); err != nil {
				// TODO
			}
		default:
			p.free = nil
			return nil
		}
	}
}

type clientconn struct {
	net.Conn
	peer      *peer
	createdAt time.Time
	usedAt    time.Time
	seq       int32 // atomic

	rTimeout time.Duration
	wTimeout time.Duration

	mu  sync.Mutex
	err error
}

func newClientconn(conn net.Conn, peer *peer) *clientconn {
	opts := peer.pool.opts
	cc := &clientconn{
		Conn:      conn,
		peer:      peer,
		createdAt: time.Now(),
		rTimeout:  opts.rTimeout,
		wTimeout:  opts.wTimeout,
	}
	return cc
}

func (c *clientconn) isExpired() bool {
	now := time.Now()
	if maxAge := c.peer.pool.opts.maxAge; maxAge > 0 && now.Sub(c.createdAt) > maxAge {
		return true
	}
	if timeout := c.peer.pool.opts.idleTimeout; timeout > 0 && now.Sub(c.usedAt) > timeout {
		return true
	}
	return false
}

func (c *clientconn) NextSequence() int32 {
	return atomic.AddInt32(&c.seq, 1)
}

func (c *clientconn) IsReused() bool {
	return !c.usedAt.IsZero()
}

func (c *clientconn) Read(b []byte) (n int, err error) {
	if c.rTimeout > 0 {
		if err = c.SetReadDeadline(time.Now().Add(c.rTimeout)); err != nil {
			return 0, err
		}
	}
	n, err = c.Conn.Read(b)
	if err != nil {
		if err == io.ErrClosedPipe || (n == 0 && err == io.EOF) {
			err = errPeerClosed
		}
		c.setError(err)
	}
	return n, err
}

func (c *clientconn) Write(b []byte) (n int, err error) {
	if c.wTimeout > 0 {
		if err = c.SetWriteDeadline(time.Now().Add(c.wTimeout)); err != nil {
			return 0, err
		}
	}
	n, err = c.Conn.Write(b)
	if err != nil {
		if err == io.ErrClosedPipe || strings.Contains(err.Error(), "broken pipe") {
			err = errPeerClosed
		}
		c.setError(err)
	}
	return n, err
}

func (c *clientconn) Close() error {
	c.setError(errConnClosed)
	return c.Conn.Close()
}

func (c *clientconn) SetDeadline(t time.Time) (err error) {
	if err = c.Conn.SetDeadline(t); err != nil {
		c.setError(err)
	}
	return
}

func (c *clientconn) SetReadDeadline(t time.Time) (err error) {
	if err = c.Conn.SetReadDeadline(t); err != nil {
		c.setError(err)
	}
	return
}

func (c *clientconn) SetWriteDeadline(t time.Time) (err error) {
	if err = c.Conn.SetWriteDeadline(t); err != nil {
		c.setError(err)
	}
	return
}

func (c *clientconn) SetTimeout(t time.Duration) (err error) {
	if err = c.SetReadTimeout(t); err == nil {
		err = c.SetWriteTimeout(t)
	}
	return err
}

func (c *clientconn) SetReadTimeout(t time.Duration) (err error) {
	if c.rTimeout != t {
		if t == 0 { // reset to no timeout
			return c.SetReadDeadline(time.Time{})
		}
		c.rTimeout = t
	}
	return nil
}

func (c *clientconn) SetWriteTimeout(t time.Duration) error {
	if c.wTimeout != t {
		if t == 0 { // reset to no timeout
			return c.SetWriteDeadline(time.Time{})
		}
		c.wTimeout = t
	}
	return nil
}

func (c *clientconn) setError(err error) {
	if err != nil {
		c.mu.Lock()
		if c.err == nil {
			c.err = err
		}
		c.mu.Unlock()
	}
}

func (c *clientconn) getError() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}
