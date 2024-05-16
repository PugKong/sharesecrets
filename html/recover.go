package html

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
)

type RecoverMiddleware struct {
	logger   *slog.Logger
	renderer *Renderer
}

func NewRecoverMiddleware(logger *slog.Logger, renderer *Renderer) *RecoverMiddleware {
	return &RecoverMiddleware{
		logger:   logger,
		renderer: renderer,
	}
}

func (m *RecoverMiddleware) Handler(next http.Handler) http.Handler {
	const stackSize = 4096

	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() { //nolint:contextcheck
			rec := recover()
			if rec == nil {
				return
			}

			err, ok := rec.(error)
			if !ok {
				err = fmt.Errorf("%v", rec) //nolint:goerr113
			}

			stack := make([]byte, stackSize)
			length := runtime.Stack(stack, false)
			stack = stack[:length]

			m.logger.LogAttrs(r.Context(), slog.LevelError, "Panic recovered",
				slog.String("error", err.Error()),
				slog.String("stack", string(stack)),
			)

			m.renderer.ServerError(r.Context(), w, err)
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
