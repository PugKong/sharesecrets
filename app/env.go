package app

import "log/slog"

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
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
