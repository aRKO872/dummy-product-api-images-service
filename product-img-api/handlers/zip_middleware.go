package handlers

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type GZipHandler struct {
}

// Header should be sent as Accept Encoding = "gzip"
func (g *GZipHandler) GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// create a zipped response
			grw := NewGZipResponseWriter(rw)
			grw.Header().Set("Content-Encoding", "gzip")
			next.ServeHTTP(grw, r)
			return
		}

		next.ServeHTTP(rw, r)
	})
}

type GZipResponseWriter struct {
	rw http.ResponseWriter
	gw *gzip.Writer
}

func NewGZipResponseWriter (rw http.ResponseWriter) *GZipResponseWriter {
	gzipWriter := gzip.NewWriter(rw)
	return &GZipResponseWriter{
		rw: rw,
		gw: gzipWriter,
	}
}

func (gwr *GZipResponseWriter) Write(d []byte) (int, error) {
	return gwr.gw.Write(d)
}

func (gwr *GZipResponseWriter) Header() http.Header {
	return gwr.rw.Header()
}

func (gwr *GZipResponseWriter) WriteHeader(statuscode int) {
	gwr.rw.WriteHeader(statuscode)
}

func (gwr *GZipResponseWriter) Flush() {
	gwr.gw.Flush()
	gwr.gw.Close()
}