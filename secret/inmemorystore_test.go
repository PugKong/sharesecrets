package secret

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInMemoryStore(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()

	t.Run("it saves, loads and removes items", func(t *testing.T) {
		store := NewInMemoryStore(logger)

		const key = "key"
		saveSecret := Secret{data: []byte("store test")}
		err := store.Save(ctx, key, saveSecret)
		require.NoError(t, err)

		loadSecret, err := store.Load(ctx, key)
		require.NoError(t, err)
		require.Equal(t, saveSecret, loadSecret)

		err = store.Remove(ctx, key)
		require.NoError(t, err)

		_, err = store.Load(ctx, key)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("remove method doesn't produce error when item not exist", func(t *testing.T) {
		store := NewInMemoryStore(logger)

		err := store.Remove(ctx, "key")
		require.NoError(t, err)
	})

	t.Run("it removes expired items", func(t *testing.T) {
		store := NewInMemoryStore(logger)

		const expiredSecretKey = "expired"
		err := store.Save(ctx, expiredSecretKey, Secret{exp: time.Now().Add(-time.Minute)})
		require.NoError(t, err)

		const activeSecretKey = "active"
		err = store.Save(ctx, activeSecretKey, Secret{exp: time.Now().Add(time.Minute)})
		require.NoError(t, err)

		store.Cleanup(ctx)

		_, err = store.Load(ctx, expiredSecretKey)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)

		_, err = store.Load(ctx, activeSecretKey)
		require.NoError(t, err)
	})
}
