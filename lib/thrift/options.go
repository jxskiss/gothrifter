package thrift

import (
	"time"
)

type Option func(options) options

type options struct {
	maxAge    time.Duration
	maxIdle   int
	maxActive int

	rTimeout    time.Duration
	wTimeout    time.Duration
	idleTimeout time.Duration

	rbufsz       int
	wbufsz       int
	maxframesize int

	nocopy  bool // TODO
	header  bool
	protoID ProtocolID
}

var DefaultOptions = options{
	maxAge:    2 * time.Second,
	maxIdle:   0, // no keepalive
	maxActive: 10000,
	rbufsz:    2048,
	wbufsz:    2048,
	//nocopy:    true,
	header:    false,
	protoID:   ProtocolIDBinary,
}

// WithMaxAge limits the age of connection in client cpool.
func WithMaxAge(t time.Duration) Option {
	return func(o options) options {
		o.maxAge = t
		return o
	}
}

// WithMaxIdle keeps n connections alive.
func WithMaxIdle(n int) Option {
	return func(o options) options {
		o.maxIdle = n
		return o
	}
}

// WithMaxActive limits the maxframesize connection
func WithMaxActive(n int) Option {
	return func(o options) options {
		o.maxActive = n
		return o
	}
}

// WithFramed enables framed transport with `maxframesize` size of single frame
func WithFramed(max int) Option {
	return func(o options) options {
		o.maxframesize = max
		return o
	}
}

func WithHeader() Option {
	return func(o options) options {
		o.header = true
		return o
	}
}

func DisableHeader() Option {
	return func(o options) options {
		o.header = false
		return o
	}
}

func WithCompact() Option {
	return func(o options) options {
		o.protoID = ProtocolIDCompact
		return o
	}
}

// WithBufferSize sets read and write buffer size for a connection
func WithBufferSize(r, w int) Option {
	return func(o options) options {
		o.rbufsz = r
		o.wbufsz = w
		return o
	}
}

// WithNoCopyReader ...
func WithNoCopyReader(b bool) Option {
	return func(o options) options {
		o.nocopy = b
		return o
	}
}

// WithTimeout ...
func WithTimeout(r, w time.Duration) Option {
	return func(o options) options {
		o.rTimeout = r
		o.wTimeout = w
		return o
	}
}

func WithIdleTimeout(timeout time.Duration) Option {
	return func(o options) options {
		o.idleTimeout = timeout
		return o
	}
}

type CallOption func(o options) options

func WithCallTimeout(r, w time.Duration) CallOption {
	return func(o options) options {
		o.rTimeout = r
		o.wTimeout = w
		return o
	}
}
