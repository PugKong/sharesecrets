package logger

import (
	"log/slog"
	"net/http"
	"time"
)

type RequestLoggerMiddleware struct {
	logger *slog.Logger
	now    func() time.Time
}

func NewRequestLoggerMiddleware(logger *slog.Logger) *RequestLoggerMiddleware {
	return &RequestLoggerMiddleware{
		logger: logger,
		now:    time.Now,
	}
}

func (m *RequestLoggerMiddleware) Handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := newCustomResponseWriter(w)

		m.logger.LogAttrs(r.Context(), slog.LevelInfo, "HTTP request accepted",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		t1 := m.now()
		next.ServeHTTP(ww, r)
		t2 := m.now()

		level := slog.LevelInfo
		if ww.statusCode >= http.StatusInternalServerError {
			level = slog.LevelError
		}

		m.logger.LogAttrs(r.Context(), level, "HTTP request handled",
			slog.Int("status", ww.statusCode),
			slog.String("duration", t2.Sub(t1).String()),
		)
	}

	return http.HandlerFunc(fn)
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
