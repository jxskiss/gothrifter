package kit

import (
	"context"
	"os"

	"github.com/go-kit/kit/log"
)

var (
	DefaultLogger log.Logger
)

func init() {
	DefaultLogger = log.NewLogfmtLogger(os.Stderr)
}

func LogID(ctx context.Context) string {
	return strFromCtx(ctx, LOGID)
}

func Caller(ctx context.Context) string {
	return strFromCtx(ctx, CALLER)
}

func Env(ctx context.Context) string {
	return strFromCtx(ctx, ENV)
}

func Service(ctx context.Context) string {
	return strFromCtx(ctx, SERVICE)
}

func Method(ctx context.Context) string {
	return strFromCtx(ctx, METHOD)
}

func RemoteAddr(ctx context.Context) string {
	return strFromCtx(ctx, ADDR)
}

func strFromCtx(ctx context.Context, key interface{}) string {
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}

func strDefault(str, default_ string) string {
	if str != "" {
		return str
	}
	return default_
}
