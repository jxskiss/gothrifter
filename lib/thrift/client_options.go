package thrift

import "time"

type ClientOption func(opts clientOpts) clientOpts

type clientOpts struct {
	maxAge    time.Duration
	maxIdle   int
	maxActive int

	rTimeout    time.Duration
	wTimeout    time.Duration
	idleTimeout time.Duration

	tFactory TransportFactory
}

func WithMaxAge(t time.Duration) ClientOption {
	return func(opts clientOpts) clientOpts {
		opts.maxAge = t
		return opts
	}
}

func WithMaxIdle(n int) ClientOption {
	return func(opts clientOpts) clientOpts {
		opts.maxIdle = n
		return opts
	}
}

func WithMaxActive(n int) ClientOption {
	return func(opts clientOpts) clientOpts {
		opts.maxActive = n
		return opts
	}
}

func WithTimeout(rTimeout, wTimeout time.Duration) ClientOption {
	return func(opts clientOpts) clientOpts {
		opts.rTimeout = rTimeout
		opts.wTimeout = wTimeout
		return opts
	}
}

func WithIdleTimeout(timeout time.Duration) ClientOption {
	return func(opts clientOpts) clientOpts {
		opts.idleTimeout = timeout
		return opts
	}
}

func WithTransportFactory(tFactory TransportFactory) ClientOption {
	return func(opts clientOpts) clientOpts {
		switch tFactory.(type) {
		case *transportFactory, *BufferedTransportFactory:
			// pass
		default:
			opts.tFactory = tFactory
		}
		return opts
	}
}

func defaultClientOptions() clientOpts {
	return clientOpts{
		maxAge:    2 * time.Second,
		maxIdle:   0, // no keepalive
		maxActive: 10000,
	}
}

type CallOption func(opts callOpts) callOpts

type callOpts struct {
	rTimeout time.Duration
	wTimeout time.Duration
}

func CallTimeout(timeout time.Duration) CallOption {
	return func(opts callOpts) callOpts {
		opts.rTimeout = timeout
		opts.wTimeout = timeout
		return opts
	}
}

func CallRwTimeout(rTimeout, wTimeout time.Duration) CallOption {
	return func(opts callOpts) callOpts {
		opts.rTimeout = rTimeout
		opts.wTimeout = wTimeout
		return opts
	}
}

func defaultCallOptions(cOpts clientOpts) callOpts {
	return callOpts{
		rTimeout: cOpts.rTimeout,
		wTimeout: cOpts.wTimeout,
	}
}
