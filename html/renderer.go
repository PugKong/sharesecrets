package html

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
)

type Renderer struct {
	logger *slog.Logger
}

func NewRenderer(logger *slog.Logger) *Renderer {
	return &Renderer{logger: logger}
}

func (r *Renderer) Component(ctx context.Context, w http.ResponseWriter, status int, component templ.Component) {
	w.WriteHeader(status)

	if err := component.Render(ctx, w); err != nil {
		r.logger.LogAttrs(ctx, slog.LevelError, "Failed to render component", slog.String("error", err.Error()))
	}
}

func (r *Renderer) UserError(ctx context.Context, w http.ResponseWriter, err error) {
	r.logger.LogAttrs(ctx, slog.LevelInfo, "User input error", slog.String("error", err.Error()))

	w.WriteHeader(http.StatusBadRequest)
	if err := UserError().Render(ctx, w); err != nil {
		r.logger.LogAttrs(ctx, slog.LevelError, "Failed to render user error page", slog.String("error", err.Error()))
	}
}

func (r *Renderer) ServerError(ctx context.Context, w http.ResponseWriter, err error) {
	r.logger.LogAttrs(ctx, slog.LevelError, "Server error", slog.String("error", err.Error()))

	w.WriteHeader(http.StatusInternalServerError)
	if err := ServerError().Render(ctx, w); err != nil {
		r.logger.LogAttrs(ctx, slog.LevelError, "Failed to render server error page", slog.String("error", err.Error()))
	}
}
