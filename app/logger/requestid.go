package logger

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func NewRequestIDMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	const header = "X-Request-Id"

	randomID := func(ctx context.Context) string {
		const randomBytes = 8
		bytes := make([]byte, randomBytes)
		if _, err := rand.Read(bytes); err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to generate request ID", slog.String("error", err.Error()))
		}

		return fmt.Sprintf("%x-%x", time.Now().Unix(), bytes)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(header)
			if requestID == "" {
				requestID = randomID(r.Context())
			}

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			req := r.WithContext(ctx)

			next.ServeHTTP(w, req)
		}

		return http.HandlerFunc(fn)
	}
}
