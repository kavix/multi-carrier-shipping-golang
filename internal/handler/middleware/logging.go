package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriterWrapper wraps standard http.ResponseWriter to capture status code.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// RequestLogger returns a middleware that logs incoming HTTP requests using slog.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status code if WriteHeader is not called
			}

			next.ServeHTTP(wrapper, r)

			duration := time.Since(start)

			logger.Info("HTTP request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", wrapper.statusCode),
				slog.String("duration", duration.String()),
				slog.String("remote_ip", r.RemoteAddr),
			)
		})
	}
}
