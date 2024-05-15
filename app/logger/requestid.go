package logger

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type RequestIDMiddleware struct {
	logger *slog.Logger
	rand   func([]byte) (int, error)
	now    func() time.Time
}

func NewRequestIDMiddleware(logger *slog.Logger) *RequestIDMiddleware {
	return &RequestIDMiddleware{
		logger: logger,
		rand:   rand.Read,
		now:    time.Now,
	}
}

const requestIDHeader = "X-Request-Id"

func (m *RequestIDMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = m.randomID(r.Context())
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}

func (m *RequestIDMiddleware) randomID(ctx context.Context) string {
	const randomBytes = 8
	bytes := make([]byte, randomBytes)
	if _, err := m.rand(bytes); err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "Failed to generate request ID", slog.String("error", err.Error()))
	}

	return fmt.Sprintf("%x-%x", m.now().Unix(), bytes)
}
