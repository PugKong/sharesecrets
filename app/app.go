package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/pugkong/sharesecrets/app/logger"
	"github.com/pugkong/sharesecrets/secret"
)

type App struct {
	env *env
}

func New(getenv func(string) string) *App {
	return &App{
		env: newEnv(getenv),
	}
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)

	logger := logger.New(os.Stderr, a.env.LogLevel(), a.env.TintedLogger())

	slog.SetLogLoggerLevel(slog.LevelError)
	slog.SetDefault(logger.With(slog.String("layer", "fallback")))

	encryptor := secret.NewSecretboxEncryptor(logger.With(slog.String("layer", "encryptor")))
	store := secret.NewInMemoryStore(logger.With(slog.String("layer", "store")))
	secrets := secret.NewService(
		logger.With(slog.String("layer", "service")),
		encryptor,
		store,
		time.Minute,
		time.Now,
	)

	server := newServer(logger.With("layer", "http"), secrets, a.env.ListenAddr())

	var services sync.WaitGroup
	start := func(serve func(ctx context.Context) error) {
		services.Add(1)

		go func() {
			defer services.Done()

			if err := serve(ctx); !errors.Is(err, context.Canceled) {
				cancel(err)
			}
		}()
	}

	start(server.Run)
	start(secrets.CleanupLoop)

	logger.Info("Application started")
	services.Wait()
	logger.Info("Application stopped")

	return fmt.Errorf("app run: %w", context.Cause(ctx))
}
