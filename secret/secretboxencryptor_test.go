package secret

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretboxEncryptor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	encryptor := NewSecretboxEncryptor(logger)
	ctx := context.Background()

	const (
		passpharse = "passpharse"
		data       = "data"
	)

	encrypted, err := encryptor.Encrypt(ctx, passpharse, data)
	require.NoError(t, err)

	decrypted, err := encryptor.Decrypt(ctx, passpharse, encrypted)
	require.NoError(t, err)
	require.Equal(t, data, decrypted)

	_, err = encryptor.Decrypt(ctx, passpharse+passpharse, encrypted)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidPassphrase)
}
