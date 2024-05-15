package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"

	"github.com/lmittmann/tint"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	normalHandler := &slog.JSONHandler{}
	tintedHandler := tint.NewHandler(nil, nil)

	tests := map[string]struct {
		level                slog.Level
		tinted               bool
		expectedHandler      slog.Handler
		expectedMessageCount int
	}{
		"it creates json logger with debug level": {
			level:                slog.LevelDebug,
			tinted:               false,
			expectedHandler:      normalHandler,
			expectedMessageCount: 4,
		},
		"it creates json logger with info level": {
			level:                slog.LevelInfo,
			tinted:               false,
			expectedHandler:      normalHandler,
			expectedMessageCount: 3,
		},
		"it creates json logger with warn level": {
			level:                slog.LevelWarn,
			tinted:               false,
			expectedHandler:      normalHandler,
			expectedMessageCount: 2,
		},
		"it creates json logger with error level": {
			level:                slog.LevelError,
			tinted:               false,
			expectedHandler:      normalHandler,
			expectedMessageCount: 1,
		},
		"it creates tinted logger with debug level": {
			level:                slog.LevelDebug,
			tinted:               true,
			expectedHandler:      tintedHandler,
			expectedMessageCount: 4,
		},
		"it creates tinted logger with info level": {
			level:                slog.LevelInfo,
			tinted:               true,
			expectedHandler:      tintedHandler,
			expectedMessageCount: 3,
		},
		"it creates tinted logger with warn level": {
			level:                slog.LevelWarn,
			tinted:               true,
			expectedHandler:      tintedHandler,
			expectedMessageCount: 2,
		},
		"it creates tinted logger with error level": {
			level:                slog.LevelError,
			tinted:               true,
			expectedHandler:      tintedHandler,
			expectedMessageCount: 1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			out := &bytes.Buffer{}

			logger := New(out, test.level, test.tinted)

			require.NotNil(t, logger)
			require.IsType(t, &handlerWrapper{}, logger.Handler())
			require.IsType(t, test.expectedHandler, logger.Handler().(*handlerWrapper).Handler) //nolint:forcetypeassert

			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")

			var logged []string
			for _, s := range strings.Split(out.String(), "\n") {
				if s != "" {
					logged = append(logged, s)
				}
			}

			require.Len(t, logged, test.expectedMessageCount)
		})
	}
}

func TestHandlerWrapper(t *testing.T) {
	newLogger := func() (*slog.Logger, func(t *testing.T) []map[string]any) {
		t.Helper()

		buf := &bytes.Buffer{}
		handler := slog.NewJSONHandler(buf, nil)
		wrapper := &handlerWrapper{handler}
		logger := slog.New(wrapper)

		output := func(t *testing.T) []map[string]any {
			t.Helper()

			var ms []map[string]any
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

		return logger, output
	}

	t.Run("it passes std tests", func(t *testing.T) {
		logger, output := newLogger()
		err := slogtest.TestHandler(logger.Handler(), func() []map[string]any { return output(t) })
		require.NoError(t, err)
	})

	t.Run("it adds requestID attr", func(t *testing.T) {
		logger, output := newLogger()

		const requestID = "42"
		ctx := context.WithValue(context.Background(), requestIDKey, requestID)
		logger.InfoContext(ctx, "test")

		out := output(t)
		require.Len(t, out, 1)
		require.Equal(t, requestID, out[0][requestIDAttr])
	})

	t.Run("it wraps Handler.WithAttrs method", func(t *testing.T) {
		logger, output := newLogger()

		const attrName = "answer"
		const attrValue = "42"
		logger = logger.With(slog.String(attrName, attrValue))
		logger.Info("test")

		out := output(t)
		require.Len(t, out, 1)
		require.Equal(t, attrValue, out[0][attrName])
		require.IsType(t, &handlerWrapper{}, logger.Handler())
	})

	t.Run("it wraps Handler.WithGroup method", func(t *testing.T) {
		logger, output := newLogger()

		const groupName = "group"
		const attrName = "answer"
		const attrValue = "42"
		logger = logger.WithGroup(groupName).With(slog.String(attrName, attrValue))
		logger.Info("test")

		out := output(t)
		require.Len(t, out, 1)
		require.Equal(t, map[string]any{attrName: attrValue}, out[0][groupName])
		require.IsType(t, &handlerWrapper{}, logger.Handler())
	})
}
