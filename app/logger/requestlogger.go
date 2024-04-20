package logger

import (
	"log/slog"
	"net/http"
	"time"
)

func NewRequestLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := newCustomResponseWriter(w)

			logger.LogAttrs(r.Context(), slog.LevelInfo, "HTTP request accepted",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			t1 := time.Now()
			next.ServeHTTP(ww, r)
			t2 := time.Now()

			duration := t2.Sub(t1)

			level := slog.LevelInfo
			if ww.statusCode >= http.StatusInternalServerError {
				level = slog.LevelError
			}

			logger.LogAttrs(r.Context(), level, "HTTP request handled",
				slog.Int("status", ww.statusCode),
				slog.String("duration", duration.String()),
			)
		}

		return http.HandlerFunc(fn)
	}
}

type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newCustomResponseWriter(w http.ResponseWriter) *customResponseWriter {
	return &customResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
