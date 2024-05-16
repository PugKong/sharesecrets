package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pugkong/sharesecrets/loggertest"
	"github.com/stretchr/testify/require"
)

func TestRequestLoggerMiddleware(t *testing.T) {
	t.Run("it logs success requests", func(t *testing.T) {
		logger, output := loggertest.New()
		middleware := NewRequestLoggerMiddleware(logger)
		middleware.now = func() time.Time { return time.Time{} }
		handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		middleware.Handler(handler).ServeHTTP(w, r)

		require.Equal(
			t,
			[]map[string]any{
				{"level": "INFO", "method": "GET", "msg": "HTTP request accepted", "path": "/"},
				{"level": "INFO", "msg": "HTTP request handled", "status": float64(200), "duration": "0s"},
			},
			output(t),
		)
	})

	t.Run("it logs failure requests", func(t *testing.T) {
		logger, output := loggertest.New()
		middleware := NewRequestLoggerMiddleware(logger)
		middleware.now = func() time.Time { return time.Time{} }
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		middleware.Handler(handler).ServeHTTP(w, r)

		require.Equal(
			t,
			[]map[string]any{
				{"level": "INFO", "method": "GET", "msg": "HTTP request accepted", "path": "/"},
				{"level": "ERROR", "msg": "HTTP request handled", "status": float64(500), "duration": "0s"},
			},
			output(t),
		)
	})

	t.Run("it measures time", func(t *testing.T) {
		logger, output := loggertest.New()
		middleware := NewRequestLoggerMiddleware(logger)
		now := time.Now()
		middleware.now = func() time.Time {
			now = now.Add(10 * time.Millisecond)

			return now
		}
		handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		middleware.Handler(handler).ServeHTTP(w, r)

		require.Equal(
			t,
			[]map[string]any{
				{"level": "INFO", "method": "GET", "msg": "HTTP request accepted", "path": "/"},
				{"level": "INFO", "msg": "HTTP request handled", "status": float64(200), "duration": "10ms"},
			},
			output(t),
		)
	})
}
