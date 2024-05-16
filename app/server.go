package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/pugkong/sharesecrets/html"
	"github.com/pugkong/sharesecrets/logger"
	"github.com/pugkong/sharesecrets/secret"
)

type server struct {
	logger  *slog.Logger
	secrets *secret.Service
	server  *http.Server
}

func newServer(logger *slog.Logger, secrets *secret.Service, listen string) *server {
	return &server{
		logger:  logger,
		secrets: secrets,
		server: &http.Server{
			ReadHeaderTimeout: time.Second,
			Addr:              listen,
		},
	}
}

func (s *server) Init(ctx context.Context) error {
	assets, err := html.MakeAssets()
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "HTTP server assets initialization error", slog.String("error", err.Error()))

		return fmt.Errorf("assets initialization: %w", err)
	}

	renderer := html.NewRenderer(s.logger)
	secretHandler := secret.NewHandler(s.secrets, renderer)

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", secretHandler.Share)
	mux.HandleFunc("/{key}", secretHandler.Open)

	var handler http.Handler = mux
	handler = html.NewRecoverMiddleware(s.logger, renderer).Handler(handler)
	handler = html.NewCSRFMiddleware(s.logger, renderer)(handler)
	handler = html.NewAssetsMiddleware(s.logger, assets)(handler)
	handler = html.NewParseFormMiddleware(renderer)(handler)
	handler = logger.NewRequestLoggerMiddleware(s.logger).Handler(handler)
	handler = logger.NewRequestIDMiddleware(s.logger).Handler(handler)
	s.server.Handler = handler

	return nil
}

func (s *server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		s.logger.InfoContext(ctx, "HTTP server started on "+s.server.Addr)
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			s.logger.LogAttrs(ctx, slog.LevelError, "HTTP server serve error", slog.String("error", err.Error()))
			cancel(err)
		}
	}()
	<-ctx.Done()

	serveErr := fmt.Errorf("http serve: %w", context.Cause(ctx))

	s.logger.InfoContext(ctx, "Shutting down HTTP server")
	if err := s.server.Shutdown(context.WithoutCancel(ctx)); err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "HTTP server shutdown error", slog.String("error", err.Error()))

		return errors.Join(fmt.Errorf("http shutdown: %w", err), serveErr)
	}
	s.logger.InfoContext(ctx, "HTTP server stopped")

	return serveErr
}
