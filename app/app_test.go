package app

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	t.Run("it stops on service error", func(t *testing.T) {
		addr, free := occupyRandomPort(t)
		defer free()

		env := mapenv(map[string]string{
			"APP_LISTEN":     addr,
			"APP_LOG_OUTPUT": "discard",
		})
		app := New(env)

		err := app.Run(context.Background())
		var target *net.OpError
		require.ErrorAs(t, err, &target)
	})

	t.Run("it stops on context cancelation", func(t *testing.T) {
		env := mapenv(map[string]string{"APP_LOG_OUTPUT": "discard"})
		app := New(env)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := app.Run(ctx)
		require.ErrorIs(t, err, context.Canceled)
	})
}

func occupyRandomPort(t *testing.T) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	return listener.Addr().String(), func() {
		if err := listener.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
