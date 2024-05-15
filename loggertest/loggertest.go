package loggertest

import (
	"bytes"
	"encoding/json"
	"log/slog"
)

type OutputFunc func(t TestingT) []map[string]any

func New() (*slog.Logger, OutputFunc) {
	return NewWithHandlerWrapper(func(h slog.Handler) slog.Handler { return h })
}

func NewWithHandlerWrapper(wrap func(slog.Handler) slog.Handler) (*slog.Logger, OutputFunc) {
	buf := &bytes.Buffer{}
	handler := slog.NewJSONHandler(buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
			if attr.Key != "time" {
				return attr
			}

			return slog.Attr{}
		},
	})
	logger := slog.New(wrap(handler))

	output := func(t TestingT) []map[string]any { return ParseJSON(t, buf) }

	return logger, output
}

type TestingT interface {
	Helper()
	Fatal(args ...any)
}

func ParseJSON(t TestingT, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var ms []map[string]any //nolint:prealloc
	for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}

		var m map[string]any
		if err := json.Unmarshal(line, &m); err != nil {
			t.Fatal(err)
		}
		ms = append(ms, m)
	}

	return ms
}
