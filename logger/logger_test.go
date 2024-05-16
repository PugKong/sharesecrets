package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"

	"github.com/lmittmann/tint"
	"github.com/pugkong/sharesecrets/loggertest"
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
	t.Run("it passes std tests", func(t *testing.T) {
		buf := &bytes.Buffer{}
		handler := handlerWrapper{slog.NewJSONHandler(buf, nil)}

		err := slogtest.TestHandler(&handler, func() []map[string]any { return loggertest.ParseJSON(t, buf) })
		require.NoError(t, err)
	})

	wrap := func(h slog.Handler) slog.Handler { return &handlerWrapper{h} }

	t.Run("it adds requestID attr", func(t *testing.T) {
		logger, output := loggertest.NewWithHandlerWrapper(wrap)

		const requestID = "42"
		ctx := context.WithValue(context.Background(), requestIDKey, requestID)
		logger.InfoContext(ctx, "test")

		out := output(t)
		require.Len(t, out, 1)
		require.Equal(t, requestID, out[0][requestIDAttr])
	})

	t.Run("it wraps Handler.WithAttrs method", func(t *testing.T) {
		logger, output := loggertest.NewWithHandlerWrapper(wrap)

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
		logger, output := loggertest.NewWithHandlerWrapper(wrap)

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
