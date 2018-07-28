package thrift

import (
	"context"
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

type Pool interface {
	Take(ctx context.Context, address string) (Conn, error)
}

type Conn interface {
	net.Conn
	SetTimeout(t time.Duration) error
	SetReadTimeout(t time.Duration) error
	SetWriteTimeout(t time.Duration) error

	NextSequence() int32
	IsReused() bool
	Put() error
}

type pool struct {
	dial  Dialer
	peers sync.Map

	opts clientOpts
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

	pp = &peer{
		pool:    p,
		address: address,
	}
	if p.opts.maxIdle > 0 {
		pp.free = make(chan *clientconn, p.opts.maxIdle)
	}
	if x, loaded := p.peers.LoadOrStore(address, pp); loaded {
		pp = x.(*peer)
	}
	return pp.take(ctx)
}

type peer struct {
	pool    *pool
	address string

	free   chan *clientconn
	active int32 // atomic
}

func (p *peer) take(ctx context.Context) (Conn, error) {
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
		if conn.Conn != nil {
			return conn.Conn.Close()
		}
	}
	if p.pool.opts.idleTimeout > 0 {
		conn.idleTimeoutAt = time.Now().Add(p.pool.opts.idleTimeout)
	}
	select {
	case p.free <- conn:
		return nil
	default:
		atomic.AddInt32(&p.active, -1)
		return conn.Conn.Close()
	}
}

type clientconn struct {
	net.Conn
	peer *peer

	maxAgeAt      time.Time
	idleTimeoutAt time.Time
	rTimeout      time.Duration
	wTimeout      time.Duration

	mu  sync.Mutex
	err error

	isReused int32 // atomic
	seq      int32 // atomic
}

func newClientconn(conn net.Conn, peer *peer) *clientconn {
	opts := peer.pool.opts
	cc := &clientconn{
		Conn:     conn,
		peer:     peer,
		rTimeout: opts.rTimeout,
		wTimeout: opts.wTimeout,
	}
	if opts.maxAge > 0 {
		cc.maxAgeAt = time.Now().Add(opts.maxAge)
	}
	return cc
}

func (c *clientconn) isExpired() bool {
	now := time.Now()
	if !c.maxAgeAt.IsZero() && now.After(c.maxAgeAt) {
		return true
	}
	if !c.idleTimeoutAt.IsZero() && now.After(c.idleTimeoutAt) {
		return true
	}
	return false
}

func (c *clientconn) NextSequence() int32 {
	return atomic.AddInt32(&c.seq, 1)
}

func (c *clientconn) IsReused() bool {
	return atomic.LoadInt32(&c.isReused) == 1
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
	}
	return n, c.setError(err)
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
	}
	return n, c.setError(err)
}

func (c *clientconn) Put() error {
	atomic.StoreInt32(&c.isReused, 1)
	return c.peer.put(c)
}

func (c *clientconn) Close() error {
	return c.setError(c.Conn.Close())
}

func (c *clientconn) SetDeadline(t time.Time) (err error) {
	return c.setError(c.Conn.SetDeadline(t))
}

func (c *clientconn) SetReadDeadline(t time.Time) (err error) {
	return c.setError(c.Conn.SetReadDeadline(t))
}

func (c *clientconn) SetWriteDeadline(t time.Time) (err error) {
	return c.setError(c.Conn.SetWriteDeadline(t))
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

func (c *clientconn) setError(err error) error {
	if err != nil {
		c.mu.Lock()
		c.err = err
		c.mu.Unlock()
	}
	return err
}

func (c *clientconn) getError() (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}
