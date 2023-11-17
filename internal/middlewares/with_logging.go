package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

type logger interface {
	Info(msg string)
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write writes response and counts response size.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader writes header and save status code.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging is middleware to log request method, path and response size, status code, execution time.
func WithLogging(handler http.Handler, logger logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		handler.ServeHTTP(&lw, r)
		duration := time.Since(start)

		logger.Info(fmt.Sprintf(
			"%s %s - %d %dB in %s",
			r.Method,
			r.RequestURI,
			lw.responseData.status,
			lw.responseData.size,
			duration.String()),
		)
	})
}
