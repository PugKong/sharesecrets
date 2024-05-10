package html

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
)

func NewRecoverMiddleware(logger *slog.Logger, renderer *Renderer) func(http.Handler) http.Handler {
	const stackSize = 4096

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() { //nolint:contextcheck
				rec := recover()
				if rec == nil {
					return
				}

				err, ok := rec.(error)
				if !ok {
					err = fmt.Errorf("%v", err) //nolint
				}

				stack := make([]byte, stackSize)
				length := runtime.Stack(stack, false)
				stack = stack[:length]

				logger.LogAttrs(r.Context(), slog.LevelError, "Panic recovered",
					slog.String("error", err.Error()),
					slog.String("stack", string(stack)),
				)

				renderer.ServerError(r.Context(), w, err)
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
