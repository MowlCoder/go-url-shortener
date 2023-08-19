package middlewares

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter

	contentTypes map[string]struct{}
	encodeFormat string
	w            io.Writer
}

func (cw *compressWriter) WriteHeader(code int) {
	if !cw.isCompressible() {
		cw.ResponseWriter.WriteHeader(code)
		return
	}

	cw.Header().Set("Content-Encoding", cw.encodeFormat)
	cw.ResponseWriter.WriteHeader(code)
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	if !cw.isCompressible() {
		return cw.ResponseWriter.Write(b)
	}

	return cw.writer().Write(b)
}

func (cw *compressWriter) Close() error {
	if c, ok := cw.writer().(io.WriteCloser); ok {
		return c.Close()
	}

	return errors.New("don't have close method")
}

func (cw *compressWriter) isCompressible() bool {
	contentType := cw.Header().Get("Content-Type")

	// Remove part with text encoding like text/plain;charset=utf-8
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = contentType[0:idx]
	}

	if _, ok := cw.contentTypes[contentType]; ok {
		return true
	}

	return false
}

func (cw *compressWriter) writer() io.Writer {
	if cw.isCompressible() {
		return cw.w
	} else {
		return cw.ResponseWriter
	}
}

func newGzipWriter(rw http.ResponseWriter, w io.Writer) *compressWriter {
	return &compressWriter{
		ResponseWriter: rw,
		w:              w,
		encodeFormat:   "gzip",
		contentTypes: map[string]struct{}{
			"application/json": {},
			"text/html":        {},
		},
	}
}

func GzipCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			defer gr.Close()

			r.Body = gr
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gzipWriter := newGzipWriter(w, gz)

		defer gzipWriter.Close()

		next.ServeHTTP(gzipWriter, r)
	})
}
