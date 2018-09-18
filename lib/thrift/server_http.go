package thrift

import (
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"strings"
)

type HttpProcessor interface {
	ProcessHttp(ctx context.Context, r *http.Request, w http.ResponseWriter) error
}

// NewThriftHandlerFunc is a function that create a ready to use Apache Thrift Handler function
func NewThriftHandlerFunc(processor HttpProcessor) func(w http.ResponseWriter, r *http.Request) {

	return gz(func(w http.ResponseWriter, r *http.Request) {
		// TODO: do we need to check Content-Type here ?

		ctx := r.Context()
		// TODO: protocol in context ?
		ctx = context.WithValue(ctx, remoteAddrCtxKey{}, r.RemoteAddr)
		processor.ProcessHttp(ctx, r, w)

		// TODO: flush ?

	})
}

// gz transparently compresses the HTTP response if the client supports it.
func gz(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			handler(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		handler(gzw, r)
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
