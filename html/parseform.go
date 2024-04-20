package html

import (
	"fmt"
	"net/http"
)

func NewParseFormMiddleware(renderer *Renderer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)

				return
			}

			err := r.ParseForm()
			if err == nil {
				next.ServeHTTP(w, r)

				return
			}

			renderer.UserError(r.Context(), w, fmt.Errorf("parse form: %w", err))
		}

		return http.HandlerFunc(fn)
	}
}
