package secret

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

var _ Store = &InMemoryStore{}

type InMemoryStore struct {
	lock   sync.Mutex
	logger *slog.Logger
	data   map[string]Secret
}

func NewInMemoryStore(logger *slog.Logger) *InMemoryStore {
	return &InMemoryStore{
		logger: logger,
		data:   make(map[string]Secret),
	}
}

func (s *InMemoryStore) Load(ctx context.Context, key string) (Secret, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	secret, ok := s.data[key]
	if !ok {
		s.logger.LogAttrs(ctx, slog.LevelDebug, "Secret not found", slog.String("key", key))

		return secret, ErrNotFound
	}
	s.logger.LogAttrs(ctx, slog.LevelDebug, "Secret loaded", slog.String("key", key))

	return secret, nil
}

func (s *InMemoryStore) Save(ctx context.Context, key string, secret Secret) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = secret
	s.logger.LogAttrs(ctx, slog.LevelDebug, "Secret saved",
		slog.String("key", key),
		slog.String("expireAt", secret.exp.Format(time.RFC3339)),
	)

	return nil
}

func (s *InMemoryStore) Remove(ctx context.Context, key string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
	s.logger.LogAttrs(ctx, slog.LevelDebug, "Secret removed", slog.String("key", key))

	return nil
}

func (s *InMemoryStore) Cleanup(ctx context.Context) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for key, secret := range s.data {
		if time.Now().After(secret.exp) {
			delete(s.data, key)

			s.logger.LogAttrs(ctx, slog.LevelDebug, "Expired secret removed", slog.String("key", key))
		}
	}
}
