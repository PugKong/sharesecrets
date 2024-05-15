package loggertest

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	logger, output := New()

	logger.Debug("debug message")
	logger.Info("info message")

	require.Equal(
		t,
		[]map[string]any{
			{"level": "DEBUG", "msg": "debug message"},
			{"level": "INFO", "msg": "info message"},
		},
		output(t),
	)
}

func TestNewWithHandlerWrapper(t *testing.T) {
	wrapper := &handlerWrapper{}
	logger, _ := NewWithHandlerWrapper(func(h slog.Handler) slog.Handler {
		wrapper.Handler = h

		return wrapper
	})
	logger.Info("some info")
	require.Equal(t, 1, wrapper.handled)
}

type handlerWrapper struct {
	slog.Handler
	handled int
}

func (h *handlerWrapper) Handle(context.Context, slog.Record) error {
	h.handled++

	return nil
}

func TestParseJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	_, err := buf.WriteString("something special, but not json\n")
	require.NoError(t, err)

	spy := &testingTSpy{}
	_ = ParseJSON(spy, buf)
	require.True(t, spy.helperCalled)
	require.Equal(t, "[invalid character 's' looking for beginning of value]", spy.fatalArgs)
}

type testingTSpy struct {
	helperCalled bool
	fatalArgs    string
}

func (t *testingTSpy) Helper() {
	t.helperCalled = true
}

func (t *testingTSpy) Fatal(args ...any) {
	t.fatalArgs = fmt.Sprintf("%+v", args)
}
