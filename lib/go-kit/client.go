package kit

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	"github.com/jxskiss/thriftkit/lib/thrift"
	"github.com/rs/xid"
)

type Client struct {
	caller   string
	service  string
	ifactory func(address string) (thrift.ProtocolInvoker, error)
	cfactory func(invoker thrift.Invoker) endpoint.Endpoint
	opts     []thrift.Option
	mws      []endpoint.Middleware

	chain      endpoint.Middleware
	instancer  sd.Instancer
	endpointer sd.Endpointer
	balancer   lb.Balancer
	logger     log.Logger
}

func NewClient(caller, service string, opts ...thrift.Option) *Client {
	// Always use header transport for kit client.
	opts = append(opts, thrift.WithHeader())
	kc := &Client{
		caller:   caller,
		service:  service,
		ifactory: thrift.NewProtocolInvokerFactory(thrift.StdDialer, opts...),
		opts:     opts,
		chain:    SetHeadersMiddleware,
		logger:   DefaultLogger,
	}
	return kc
}

func (kc *Client) UseFactory(factory func(invoker thrift.Invoker) endpoint.Endpoint) *Client {
	kc.cfactory = factory
	return kc
}

func (kc *Client) UseAddress(addr string) *Client {
	instancer := sd.FixedInstancer{addr}
	return kc.UseInstancer(instancer)
}

func (kc *Client) UseInstancer(instancer sd.Instancer) *Client {
	endpointer := sd.NewEndpointer(instancer, kc.endpointFactory, kc.logger)
	balancer := lb.NewRoundRobin(endpointer)
	kc.resetDiscovery(balancer, instancer, endpointer)
	return kc
}

func (kc *Client) resetDiscovery(balancer lb.Balancer, instancer sd.Instancer, endpointer sd.Endpointer) {
	type closable interface{ Close() }

	if kc.balancer != nil {
		kc.balancer = nil
	}
	if kc.endpointer != nil {
		if x, ok := kc.endpointer.(closable); ok {
			x.Close()
		}
		kc.endpointer = nil
	}
	if kc.instancer != nil {
		kc.instancer.Stop()
		kc.instancer = nil
	}
	kc.balancer, kc.instancer, kc.endpointer = balancer, instancer, endpointer
}

func (kc *Client) UseBalancer(balancer lb.Balancer) *Client {
	kc.resetDiscovery(balancer, nil, nil)
	return kc
}

func (kc *Client) UseLogger(logger log.Logger) *Client {
	kc.logger = logger
	return kc
}

func (kc *Client) UseMiddleware(mw endpoint.Middleware, more ...endpoint.Middleware) *Client {
	kc.mws = append(kc.mws, mw)
	if len(more) > 0 {
		kc.mws = append(kc.mws, more...)
	}
	chain := endpoint.Chain(kc.mws[0], kc.mws[1:]...)
	kc.chain = endpoint.Chain(chain, SetHeadersMiddleware)
	return kc
}

func (kc *Client) Call(method string, ctx context.Context, request interface{}) (interface{}, error) {
	if kc.balancer == nil {
		return nil, fmt.Errorf("address/instancer/balancer not set")
	}
	ep, err := kc.balancer.Endpoint()
	if err != nil {
		return nil, err
	}
	ctx = NewClientRpcCtx(ctx, kc.caller, kc.service, method)
	return ep(ctx, request)
}

func (kc *Client) endpointFactory(instance string) (endpoint.Endpoint, io.Closer, error) {
	invoker, err := kc.ifactory(instance)
	if err != nil {
		return nil, nil, err
	}
	ep := func(ctx context.Context, request interface{}) (interface{}, error) {
		call := kc.cfactory(invoker)
		info := getClientRpcInfo(ctx)
		info.protocol = invoker.Protocol()
		return kc.chain(call)(ctx, request)
	}
	return ep, invoker, err
}

// SetHeadersMiddleware write context information from ClientRpcInfo to request.
// Generally, this should be the last middleware of any endpoint.
func SetHeadersMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		info := getClientRpcInfo(ctx)
		p := info.protocol
		// meta info
		p.SetHeader(SERVICE, info.Service)
		p.SetHeader(METHOD, info.Method)
		p.SetHeader(CALLER, info.Caller)
		// extra info
		p.SetHeader(LOGID, info.LogID)
		p.SetHeader(ENV, info.Env)
		return next(ctx, request)
	}
}

type clientRpcInfoCtxKey struct{}

type clientRpcCtx struct {
	context.Context
	protocol *thrift.Protocol

	// meta info
	Service string // the target service name
	Method  string // the target method name
	Caller  string // this service's name

	// extra info
	LogID string // the unique request ID, derive from downstream request if available
	Env   string // this caller's ENV setting
}

func NewClientRpcCtx(ctx context.Context, caller, service, method string) *clientRpcCtx {
	info := &clientRpcCtx{
		Context: ctx,
		Caller:  caller,
		Service: service,
		Method:  method,
	}
	info.DeriveServerRpcCtx(getServerRpcInfo(ctx))
	if info.LogID == "" {
		info.LogID = xid.New().String()
	}
	// TODO: is it better to specify this in config file ?
	info.Env = os.Getenv("ENV")
	return info
}

func (r *clientRpcCtx) DeriveServerRpcCtx(info *serverRpcCtx) {
	if info == nil {
		return
	}
	if info.Service != "" {
		r.Caller = info.Service
	}
	if info.LogID != "" {
		r.LogID = info.LogID
	}
}

func (r *clientRpcCtx) Value(key interface{}) interface{} {
	if _, ok := key.(clientRpcInfoCtxKey); ok {
		return r
	}
	if k, ok := key.(string); ok {
		switch k {
		case SERVICE:
			return r.Service
		case METHOD:
			return r.Method
		case CALLER:
			return r.Caller
		case LOGID:
			return r.LogID
		case ENV:
			return r.Env
		}
	}
	return r.Context.Value(key)
}

func getClientRpcInfo(ctx context.Context) *clientRpcCtx {
	if r, ok := ctx.Value(clientRpcInfoCtxKey{}).(*clientRpcCtx); ok {
		return r
	}
	return nil
}
