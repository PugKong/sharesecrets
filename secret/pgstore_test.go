package secret

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPgStore(t *testing.T) {
	ctx := context.Background()
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:16",
			Env:          map[string]string{"POSTGRES_PASSWORD": "password"},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForExposedPort(),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Could not start postgres: %v", err)
	}
	defer func() {
		if err := postgres.Terminate(ctx); err != nil {
			t.Fatalf("Could not stop postgres: %v", err)
		}
	}()

	port, err := postgres.MappedPort(ctx, nat.Port("5432/tcp"))
	require.NoError(t, err)

	logger := slog.Default()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("postgres://postgres:password@localhost:%s/postgres", port.Port()))
	require.NoError(t, err)

	store := NewPgStore(logger, pool)
	err = store.Init(ctx)
	require.NoError(t, err)

	t.Run("it saves, loads and removes items", func(t *testing.T) {
		const key = "key"
		saveSecret := Secret{data: []byte("store test"), attempts: 3, exp: time.Now()}
		err := store.Save(ctx, key, saveSecret)
		require.NoError(t, err)

		loadSecret, err := store.Load(ctx, key)
		require.NoError(t, err)
		require.Equal(t, saveSecret.data, loadSecret.data)
		require.Equal(t, saveSecret.attempts, loadSecret.attempts)
		require.Equal(t, saveSecret.exp.Format(time.RFC3339), loadSecret.exp.Format(time.RFC3339))

		err = store.Remove(ctx, key)
		require.NoError(t, err)

		_, err = store.Load(ctx, key)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("remove method doesn't produce error when item not exist", func(t *testing.T) {
		err := store.Remove(ctx, "key")
		require.NoError(t, err)
	})

	t.Run("it removes expired items", func(t *testing.T) {
		const expiredSecretKey = "expired"
		err := store.Save(ctx, expiredSecretKey, Secret{data: []byte{}, exp: time.Now().Add(-time.Minute)})
		require.NoError(t, err)

		const activeSecretKey = "active"
		err = store.Save(ctx, activeSecretKey, Secret{data: []byte{}, exp: time.Now().Add(time.Minute)})
		require.NoError(t, err)

		err = store.Cleanup(ctx)
		require.NoError(t, err)

		_, err = store.Load(ctx, expiredSecretKey)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNotFound)

		_, err = store.Load(ctx, activeSecretKey)
		require.NoError(t, err)
	})
}
