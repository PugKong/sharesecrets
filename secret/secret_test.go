package secret

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("it stores and retrieves data", func(t *testing.T) {
		input := struct{ Passpharse, Message string }{
			Passpharse: "Passpharse",
			Message:    "Message",
		}

		encryptor := NewSecretboxEncryptor(logger)
		store := NewInMemoryStore(logger)
		service := NewService(logger, encryptor, store, time.Minute, time.Now)

		key, err := service.Store(ctx, StoreRequest{
			Passphrase: input.Passpharse,
			Message:    input.Message,
			Attempts:   1,
			ExpireAt:   time.Now().Add(time.Minute),
		})
		require.NoError(t, err)
		require.NotEmpty(t, key)

		secret, err := store.Load(ctx, key)
		require.NoError(t, err)

		message, err := encryptor.Decrypt(ctx, input.Passpharse, secret.data)
		require.NoError(t, err)
		require.Equal(t, input.Message, message)

		message, err = service.Retrieve(ctx, RetrieveRequest{
			Key:        key,
			Passphrase: input.Passpharse,
		})
		require.NoError(t, err)
		require.Equal(t, input.Message, message)

		_, err = store.Load(ctx, key)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("it returns error when secret not found", func(t *testing.T) {
		service := NewService(logger, NewSecretboxEncryptor(logger), NewInMemoryStore(logger), time.Minute, time.Now)

		_, err := service.Retrieve(ctx, RetrieveRequest{
			Key:        "not-found",
			Passphrase: "any",
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("it returns error when invalid passphrase provided", func(t *testing.T) {
		const passphrase = "passphrase"

		service := NewService(logger, NewSecretboxEncryptor(logger), NewInMemoryStore(logger), time.Minute, time.Now)

		key, err := service.Store(ctx, StoreRequest{
			Passphrase: passphrase,
			Message:    "Message",
			Attempts:   1,
			ExpireAt:   time.Now().Add(time.Minute),
		})
		require.NoError(t, err)

		_, err = service.Retrieve(ctx, RetrieveRequest{
			Key:        key,
			Passphrase: passphrase + passphrase,
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidPassphrase)
	})

	t.Run("it respects attempts limit", func(t *testing.T) {
		const passphrase = "passphrase"

		service := NewService(logger, NewSecretboxEncryptor(logger), NewInMemoryStore(logger), time.Minute, time.Now)

		key, err := service.Store(ctx, StoreRequest{
			Passphrase: passphrase,
			Message:    "Message",
			Attempts:   2,
			ExpireAt:   time.Now().Add(time.Minute),
		})
		require.NoError(t, err)

		retrieveRequest := RetrieveRequest{Key: key, Passphrase: passphrase + passphrase}
		_, err = service.Retrieve(ctx, retrieveRequest)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidPassphrase)

		_, err = service.Retrieve(ctx, retrieveRequest)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidPassphrase)

		_, err = service.Retrieve(ctx, retrieveRequest)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("it respects time limit", func(t *testing.T) {
		const passphrase = "passphrase"

		service := NewService(
			logger,
			NewSecretboxEncryptor(logger),
			NewInMemoryStore(logger),
			time.Minute,
			func() time.Time { return time.Now().Add(1 * time.Minute) },
		)

		key, err := service.Store(ctx, StoreRequest{
			Passphrase: passphrase,
			Message:    "Message",
			Attempts:   1,
			ExpireAt:   time.Now(),
		})
		require.NoError(t, err)

		_, err = service.Retrieve(ctx, RetrieveRequest{
			Key:        key,
			Passphrase: passphrase,
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrExpired)
	})
}
