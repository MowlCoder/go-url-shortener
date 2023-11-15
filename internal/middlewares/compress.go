package middlewares

import (
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"strings"
)

// CompressWriter responsible for compressing response of specified content-type.
type CompressWriter struct {
	http.ResponseWriter

	contentTypes map[string]struct{}
	encodeFormat string
	w            io.Writer
}

// WriteHeader if content type is not compressible just writing header.
// If compressible, then adding Content-Encoding header as addition.
func (cw *CompressWriter) WriteHeader(code int) {
	if !cw.isCompressible() {
		cw.ResponseWriter.WriteHeader(code)
		return
	}

	cw.Header().Set("Content-Encoding", cw.encodeFormat)
	cw.ResponseWriter.WriteHeader(code)
}

// Write writes bytes to writer. Writer is base http writer if response content-type is not compressible.
// If compressible, then writer is given io.Writer.
func (cw *CompressWriter) Write(b []byte) (int, error) {
	return cw.writer().Write(b)
}

// Close closes writer.
func (cw *CompressWriter) Close() error {
	if c, ok := cw.writer().(io.WriteCloser); ok {
		return c.Close()
	}

	return errors.New("don't have close method")
}

func (cw *CompressWriter) isCompressible() bool {
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

func (cw *CompressWriter) writer() io.Writer {
	if cw.isCompressible() {
		return cw.w
	} else {
		return cw.ResponseWriter
	}
}

func newGzipWriter(rw http.ResponseWriter, w io.Writer) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: rw,
		w:              w,
		encodeFormat:   "gzip",
		contentTypes: map[string]struct{}{
			"application/json": {},
			"text/html":        {},
		},
	}
}

type writerResetter interface {
	io.Writer
	Reset(w io.Writer)
}

// CompressMiddleware is middleware to compress request and response.
type CompressMiddleware struct {
	wr writerResetter
}

// NewCompressMiddleware is construction function to create CompressMiddleware.
func NewCompressMiddleware(writerResetter writerResetter) *CompressMiddleware {
	compressor := CompressMiddleware{}
	compressor.wr = writerResetter

	return &compressor
}

// Handler need to use as middleware.
func (compressor CompressMiddleware) Handler(next http.Handler) http.Handler {
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

		compressor.wr.Reset(w)
		gzipWriter := newGzipWriter(w, compressor.wr)

		defer gzipWriter.Close()

		next.ServeHTTP(gzipWriter, r)
	})
}
