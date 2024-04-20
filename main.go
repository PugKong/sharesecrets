package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pugkong/sharesecrets/app"
)

func main() {
	run(context.Background(), os.Getenv)
}

func run(ctx context.Context, getenv func(string) string) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := app.New(getenv)
	app.Run(ctx)
}
