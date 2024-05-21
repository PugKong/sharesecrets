package app

import (
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func mapenv(env map[string]string) func(string) string {
	return func(key string) string {
		return env[key]
	}
}

func TestEnv_ListenAddr(t *testing.T) {
	tests := map[string]struct {
		env      map[string]string
		expected string
	}{
		"default value": {
			env:      nil,
			expected: "127.0.0.1:8000",
		},
		"custom value": {
			env:      map[string]string{"APP_LISTEN": "0.0.0.0:9000"},
			expected: "0.0.0.0:9000",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env := newEnv(mapenv(test.env))

			actual := env.ListenAddr()

			require.Equal(t, test.expected, actual)
		})
	}
}

func TestEnv_TintedLogger(t *testing.T) {
	tests := map[string]struct {
		env      map[string]string
		expected bool
	}{
		"default value": {
			env:      nil,
			expected: false,
		},
		"tinted": {
			env:      map[string]string{"APP_LOGGER": "tinted"},
			expected: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env := newEnv(mapenv(test.env))

			actual := env.TintedLogger()

			require.Equal(t, test.expected, actual)
		})
	}
}

func TestEnv_LogLevel(t *testing.T) {
	tests := map[string]struct {
		env      map[string]string
		expected slog.Level
	}{
		"default value": {
			env:      nil,
			expected: slog.LevelInfo,
		},
		"debug": {
			env:      map[string]string{"APP_LOG_LEVEL": "debug"},
			expected: slog.LevelDebug,
		},
		"warn": {
			env:      map[string]string{"APP_LOG_LEVEL": "warn"},
			expected: slog.LevelWarn,
		},
		"error": {
			env:      map[string]string{"APP_LOG_LEVEL": "error"},
			expected: slog.LevelError,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env := newEnv(mapenv(test.env))

			actual := env.LogLevel()

			require.Equal(t, test.expected, actual)
		})
	}
}

func TestEnv_LogOutput(t *testing.T) {
	tests := map[string]struct {
		env      map[string]string
		expected io.Writer
	}{
		"default value": {
			env:      nil,
			expected: os.Stderr,
		},
		"stdout": {
			env:      map[string]string{"APP_LOG_OUTPUT": "stdout"},
			expected: os.Stdout,
		},
		"discard": {
			env:      map[string]string{"APP_LOG_OUTPUT": "discard"},
			expected: io.Discard,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env := newEnv(mapenv(test.env))

			actual := env.LogOutput()

			require.Equal(t, test.expected, actual)
		})
	}
}

func TestEnv_DB(t *testing.T) {
	tests := map[string]struct {
		env      map[string]string
		expected string
	}{
		"default": {
			env:      nil,
			expected: "",
		},
		"postgres": {
			env:      map[string]string{"APP_DB": "postgres://postgres:password@localhost:5432/postgres"},
			expected: "postgres://postgres:password@localhost:5432/postgres",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			env := newEnv(mapenv(test.env))

			actual := env.DB()

			require.Equal(t, test.expected, actual)
		})
	}
}
