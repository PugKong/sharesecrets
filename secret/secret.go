package secret

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrInvalidPassphrase = errors.New("invalid passphrase")
	ErrExpired           = errors.New("expired")
)

type StoreRequest struct {
	Passphrase string
	Message    string
	Attempts   int
	ExpireAt   time.Time
}

type RetrieveRequest struct {
	Key        string
	Passphrase string
}

type Secret struct {
	data     []byte
	attempts int
	exp      time.Time
}

type Encryptor interface {
	Encrypt(ctx context.Context, passpharse, message string) ([]byte, error)
	Decrypt(ctx context.Context, passphrase string, data []byte) (string, error)
}

type Store interface {
	Save(ctx context.Context, key string, secret Secret) error
	Load(ctx context.Context, key string) (Secret, error)
	Remove(ctx context.Context, key string) error
	Cleanup(ctx context.Context)
}

type Service struct {
	logger          *slog.Logger
	encryptor       Encryptor
	store           Store
	cleanupInterval time.Duration
	now             func() time.Time
}

func NewService(logger *slog.Logger, encryptor Encryptor, store Store, cleanupInterval time.Duration, now func() time.Time) *Service {
	return &Service{
		logger:          logger,
		encryptor:       encryptor,
		store:           store,
		cleanupInterval: cleanupInterval,
		now:             now,
	}
}

func (s *Service) Store(ctx context.Context, request StoreRequest) (string, error) {
	key, err := s.generateStoreKey(ctx)
	if err != nil {
		return "", err
	}

	data, err := s.encryptMessage(ctx, request.Passphrase, request.Message)
	if err != nil {
		return "", err
	}

	secret := Secret{
		data:     data,
		attempts: request.Attempts,
		exp:      request.ExpireAt,
	}
	if err := s.saveSecret(ctx, key, secret); err != nil {
		return "", err
	}

	return key, nil
}

func (s *Service) Retrieve(ctx context.Context, request RetrieveRequest) (string, error) {
	secret, err := s.loadSecret(ctx, request.Key)
	if err != nil {
		return "", err
	}

	if secret.exp.Before(s.now()) {
		s.logger.LogAttrs(ctx, slog.LevelInfo, "Loaded secret is expired", slog.String("key", request.Key))

		return "", ErrExpired
	}

	message, err := s.decryptData(ctx, request.Passphrase, secret.data)
	if err != nil && !errors.Is(err, ErrInvalidPassphrase) {
		return "", err
	}

	if err == nil {
		return message, s.removeSecret(ctx, request.Key)
	}

	secret.attempts--
	if secret.attempts == 0 {
		return message, cmp.Or(s.removeSecret(ctx, request.Key), err) //nolint:wrapcheck
	}

	return "", cmp.Or(s.saveSecret(ctx, request.Key, secret), err) //nolint:wrapcheck
}

func (s *Service) CleanupLoop(ctx context.Context) bool {
	timer := time.NewTicker(s.cleanupInterval)
	defer timer.Stop()

	s.logger.InfoContext(ctx, "Secrets cleanup loop started")
	for {
		select {
		case <-timer.C:
			s.logger.InfoContext(ctx, "Secrets cleanup started")

			start := time.Now()
			s.store.Cleanup(ctx)
			duration := time.Since(start)

			s.logger.LogAttrs(ctx, slog.LevelInfo, "Secrets cleanup completed", slog.String("duration", duration.String()))
		case <-ctx.Done():
			s.logger.InfoContext(ctx, "Secrets cleanup loop stopped")

			return true
		}
	}
}

func (s *Service) loadSecret(ctx context.Context, key string) (Secret, error) {
	logger := s.logger.With(slog.String("key", key))

	secret, err := s.store.Load(ctx, key)
	if err != nil {
		level := slog.LevelInfo
		if !errors.Is(err, ErrNotFound) {
			level = slog.LevelError
		}
		logger.LogAttrs(ctx, level, "Failed to load secret", slog.String("error", err.Error()))

		return secret, fmt.Errorf("load secret: %w", err)
	}
	logger.LogAttrs(ctx, slog.LevelInfo, "Secret loaded")

	return secret, nil
}

func (s *Service) saveSecret(ctx context.Context, key string, secret Secret) error {
	logger := s.logger.With(slog.String("key", key))

	if err := s.store.Save(ctx, key, secret); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Failed to save secret", slog.String("error", err.Error()))

		return fmt.Errorf("save secret: %w", err)
	}
	logger.LogAttrs(ctx, slog.LevelInfo, "Secret saved")

	return nil
}

func (s *Service) removeSecret(ctx context.Context, key string) error {
	logger := s.logger.With(slog.String("key", key))

	if err := s.store.Remove(ctx, key); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Failed to remove secret", slog.String("error", err.Error()))

		return fmt.Errorf("remove secret: %w", err)
	}
	logger.LogAttrs(ctx, slog.LevelInfo, "Secret removed")

	return nil
}

func (s *Service) encryptMessage(ctx context.Context, passpharse, message string) ([]byte, error) {
	bytes, err := s.encryptor.Encrypt(ctx, passpharse, message)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "Failed to encrypt message", slog.String("error", err.Error()))

		return bytes, fmt.Errorf("encrypt data: %w", err)
	}
	s.logger.InfoContext(ctx, "Message encrypted")

	return bytes, nil
}

func (s *Service) decryptData(ctx context.Context, passphrase string, data []byte) (string, error) {
	message, err := s.encryptor.Decrypt(ctx, passphrase, data)
	if err != nil {
		level := slog.LevelInfo
		if !errors.Is(err, ErrInvalidPassphrase) {
			level = slog.LevelError
		}
		s.logger.LogAttrs(ctx, level, "Failed to decrypt message", slog.String("error", err.Error()))

		return message, fmt.Errorf("decrypt data: %w", err)
	}
	s.logger.InfoContext(ctx, "Message decrypted")

	return message, nil
}

func (s *Service) generateStoreKey(ctx context.Context) (string, error) {
	const length = 16

	key := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "Failed to generate random key", slog.String("error", err.Error()))

		return "", fmt.Errorf("generate random key: %w", err)
	}
	s.logger.InfoContext(ctx, "Random key generated")

	return hex.EncodeToString(key), nil
}
