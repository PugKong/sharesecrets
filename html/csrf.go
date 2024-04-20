package html

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

const (
	csrfKey        ctxKey = "csrf"
	csrfHeaderName string = "X-CSRF-Token"
)

var errInvalidCSRFToken = errors.New("invalid csrf token")

func NewCSRFMiddleware(logger *slog.Logger, renderer *Renderer) func(http.Handler) http.Handler {
	const cookieName = "csrf"

	generateToken := func(ctx context.Context) string {
		const tokenBytes = 16

		bytes := make([]byte, tokenBytes)
		if _, err := rand.Read(bytes); err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to generate csrf token", slog.String("error", err.Error()))
		}

		return hex.EncodeToString(bytes)
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead || r.Method == http.MethodOptions || r.Method == http.MethodTrace {
				next.ServeHTTP(w, r)

				return
			}

			cookie, err := r.Cookie(cookieName)
			if r.Method == http.MethodGet && errors.Is(err, http.ErrNoCookie) {
				cookie = &http.Cookie{
					Name:     cookieName,
					Value:    generateToken(r.Context()),
					HttpOnly: true,
				}
			}

			if r.Method != http.MethodGet && subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(r.Header.Get(csrfHeaderName))) == 0 {
				renderer.UserError(r.Context(), w, errInvalidCSRFToken)

				return
			}

			cookie.Expires = time.Now().Add(time.Hour)
			http.SetCookie(w, cookie)

			ctx := context.WithValue(r.Context(), csrfKey, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
