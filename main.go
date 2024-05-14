package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/pugkong/sharesecrets/app"
)

func main() {
	os.Exit(run(context.Background(), os.Getenv))
}

func run(ctx context.Context, getenv func(string) string) int {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := app.New(getenv)
	if err := app.Run(ctx); !errors.Is(err, context.Canceled) {
		return 1
	}

	return 0
}
