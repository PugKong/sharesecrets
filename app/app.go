package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pugkong/sharesecrets/logger"
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
	logger := logger.New(a.env.LogOutput(), a.env.LogLevel(), a.env.TintedLogger())

	slog.SetLogLoggerLevel(slog.LevelError)
	slog.SetDefault(logger.With(slog.String("layer", "fallback")))

	var pool *pgxpool.Pool
	if strings.HasPrefix(a.env.DB(), "postgres") {
		var err error
		pool, err = pgxpool.New(ctx, a.env.DB())
		if err != nil {
			logger.LogAttrs(ctx, slog.LevelError, "Failed to initialize postgres pool",
				slog.String("error", err.Error()),
			)

			return fmt.Errorf("postgres pool initialization: %w", err)
		}
		defer pool.Close()

		logger.InfoContext(ctx, "Using postgres storage")
	} else {
		logger.InfoContext(ctx, "Using in-memory storage")
	}

	secrets, err := a.makeSecretsService(ctx, logger, pool)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Failed to initialize secrets service", slog.String("error", err.Error()))

		return err
	}

	server := newServer(logger.With("layer", "http"), secrets, a.env.ListenAddr())
	if err := server.Init(ctx); err != nil {
		return err
	}

	ctx, cancel := context.WithCancelCause(ctx)
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

func (a *App) makeSecretsService(ctx context.Context, logger *slog.Logger, pool *pgxpool.Pool) (*secret.Service, error) {
	encryptor := secret.NewSecretboxEncryptor(logger.With(slog.String("layer", "encryptor")))

	var store secret.Store
	if pool != nil {
		s := secret.NewPgStore(pool)
		if err := s.Init(ctx); err != nil {
			return nil, fmt.Errorf("secret pg store initialization: %w", err)
		}

		store = s
	} else {
		store = secret.NewInMemoryStore(logger.With(slog.String("layer", "store")))
	}

	return secret.NewService(
		logger.With(slog.String("layer", "service")),
		encryptor,
		store,
		time.Minute,
		time.Now,
	), nil
}
