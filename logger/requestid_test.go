package logger

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/pugkong/sharesecrets/loggertest"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("it uses X-Request-Id header", func(t *testing.T) {
		middleware := NewRequestIDMiddleware(nil)

		const expectedID = "42"
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Add(requestIDHeader, expectedID)
		handler := newRequestIDHeaderSpyHandler()
		middleware.Handler(handler).ServeHTTP(nil, r)

		require.Equal(t, expectedID, handler.RequestID)
	})

	t.Run("it generates request id when X-Request-Id header is missing", func(t *testing.T) {
		middleware := NewRequestIDMiddleware(nil)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		handler := newRequestIDHeaderSpyHandler()
		wrapped := middleware.Handler(handler)
		wrapped.ServeHTTP(nil, r)

		require.Regexp(t, regexp.MustCompile("^[a-z0-9]{8}-[a-z0-9]{16}$"), handler.RequestID)

		middleware.now = func() time.Time { return time.Date(2007, time.January, 1, 0, 0, 0, 0, time.UTC) }
		middleware.rand = func(b []byte) (int, error) {
			for i := range b {
				b[i] = byte(i)
			}

			return len(b), nil
		}
		wrapped.ServeHTTP(nil, r)

		require.Equal(t, "45984f00-0001020304050607", handler.RequestID)
	})

	t.Run("it handles request id generation error", func(t *testing.T) {
		logger, output := loggertest.New()
		middleware := NewRequestIDMiddleware(logger)
		middleware.rand = func([]byte) (int, error) { return 0, io.EOF }
		middleware.now = func() time.Time { return time.Date(2007, time.January, 1, 0, 0, 0, 0, time.UTC) }

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		handler := newRequestIDHeaderSpyHandler()
		middleware.Handler(handler).ServeHTTP(nil, r)

		require.Equal(t, "45984f00-0000000000000000", handler.RequestID)
		require.Equal(
			t,
			[]map[string]any{
				{"level": "ERROR", "msg": "Failed to generate request ID", "error": io.EOF.Error()},
			},
			output(t),
		)
	})
}

type requestIDHeaderSpyHandler struct {
	RequestID string
}

func newRequestIDHeaderSpyHandler() *requestIDHeaderSpyHandler {
	return &requestIDHeaderSpyHandler{}
}

func (h *requestIDHeaderSpyHandler) ServeHTTP(_ http.ResponseWriter, r *http.Request) {
	h.RequestID, _ = r.Context().Value(requestIDKey).(string)
}
