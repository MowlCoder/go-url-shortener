package middlewares

import (
	"fmt"
	"net/http"
	"time"
)

type Logger interface {
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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(handler http.Handler, logger Logger) http.Handler {
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
