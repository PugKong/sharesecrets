package logger

import (
	"context"
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

func New(out io.Writer, level slog.Level, tinted bool) *slog.Logger {
	var handler slog.Handler
	if tinted {
		handler = tint.NewHandler(out, &tint.Options{Level: level})
	} else {
		handler = slog.NewJSONHandler(out, &slog.HandlerOptions{Level: level})
	}

	return slog.New(&handlerWrapper{handler})
}

type ctxKey string

const requestIDKey ctxKey = "requestID"

type handlerWrapper struct {
	slog.Handler
}

func (h *handlerWrapper) Handle(ctx context.Context, record slog.Record) error {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		record.AddAttrs(slog.String("requestID", requestID))
	}

	return h.Handler.Handle(ctx, record) //nolint:wrapcheck
}

func (h *handlerWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	handler := h.Handler.WithAttrs(attrs)

	return &handlerWrapper{handler}
}

func (h *handlerWrapper) WithGroup(name string) slog.Handler {
	handler := h.Handler.WithGroup(name)

	return &handlerWrapper{handler}
}
