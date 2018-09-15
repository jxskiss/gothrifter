package kit

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
	"github.com/rs/xid"
)

const (
	LOGID   = "LOGID"   // unique request ID
	CALLER  = "CALLER"  // upstream service name
	ENV     = "ENV"     // upstream service env
	ADDR    = "ADDR"    // upstream service ip address
	METHOD  = "METHOD"  // the method being called
	SERVICE = "SERVICE" // the service being called
)

func MW(next endpoint.Endpoint, logger log.Logger) endpoint.Endpoint {
	return endpoint.Chain(
		ContextMiddleware,
		mkLoggingMiddleware(logger),
		mkRecoverMiddleware(logger),
	)(next)
}

// ContextMiddleware populate header information from request into ServerRpcInfo context.
// Generally, this should be the first middleware of any endpoint.
func ContextMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		p := thrift.ProtocolFromCtx(ctx)
		// service, method is already filled when creating the ServerRpcInfo
		info := getServerRpcInfo(ctx)
		info.RemoteAddr = thrift.RemoteAddrFromCtx(ctx)

		// read request headers
		info.headers = p.ReadHeaders()
		if info.LogID = info.headers[LOGID]; info.LogID == "" {
			info.LogID = xid.New().String()
		}
		info.Caller = info.headers[CALLER]
		info.Env = info.headers[ENV]
		return next(ctx, request)
	}
}

func mkLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			logger = level.Info(logger)
			defer func(begin time.Time) {
				r := getServerRpcInfo(ctx)
				t := time.Now()
				var cost, errstr string
				cost = strconv.FormatInt(t.Sub(begin).Nanoseconds()/1000, 10) // us
				if err != nil {
					errstr = err.Error()
				}
				logger.Log(
					"time", t.UTC().Format(time.RFC3339Nano),
					"logid", r.LogID,
					"service", r.Service,
					"method", r.Method,
					"caller", r.Caller,
					"addr", r.RemoteAddr,
					"env", r.Env,
					"cost", cost,
					"err", errstr,
				)
			}(time.Now())
			return next(ctx, request)
		}
	}
}

func mkRecoverMiddleware(logger log.Logger) endpoint.Middleware {
	logger = level.Error(logger)
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					const size = 64 << 10
					buf := make([]byte, size)
					buf = buf[:runtime.Stack(buf, false)]
					logger.Log("msg", "panic in handler", "stack", buf)
					if e, ok := r.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("%v", r)
					}
				}
			}()
			return next(ctx, request)
		}
	}
}

type serverRpcInfoCtxKey struct{}

type serverRpcCtx struct {
	context.Context

	// meta info
	Service string // this service name
	Method  string // the method being called
	Caller  string // the upstream service name

	// extra info
	LogID      string // the unique request ID
	Env        string // the client's ENV setting
	RemoteAddr string // the client's ip address

	// request headers, available for header protocol
	headers map[string]string
}

func NewServerRpcCtx(ctx context.Context, service, method string) *serverRpcCtx {
	return &serverRpcCtx{
		Context: ctx,
		Service: service,
		Method:  method,
	}
}

func (r *serverRpcCtx) Value(key interface{}) interface{} {
	if _, ok := key.(serverRpcInfoCtxKey); ok {
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
		case ADDR:
			return r.RemoteAddr
		}
	}
	return r.Context.Value(key)
}

func getServerRpcInfo(ctx context.Context) *serverRpcCtx {
	if r, ok := ctx.Value(serverRpcInfoCtxKey{}).(*serverRpcCtx); ok {
		return r
	}
	return nil
}
