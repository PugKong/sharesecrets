package app

import (
	"io"
	"log/slog"
	"os"
)

type env struct {
	getenv func(string) string
}

func newEnv(getenv func(string) string) *env {
	return &env{getenv: getenv}
}

func (e *env) ListenAddr() string {
	if addr := e.getenv("APP_LISTEN"); addr != "" {
		return addr
	}

	return "127.0.0.1:8000"
}

func (e *env) TintedLogger() bool {
	return e.getenv("APP_LOGGER") == "tinted"
}

func (e *env) LogLevel() slog.Level {
	switch e.getenv("APP_LOG_LEVEL") {
	default:
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	}
}

func (e *env) LogOutput() io.Writer {
	switch e.getenv("APP_LOG_OUTPUT") {
	default:
		return os.Stderr
	case "stdout":
		return os.Stdout
	case "discard":
		return io.Discard
	}
}

func (e *env) DB() string {
	return e.getenv("APP_DB")
}
