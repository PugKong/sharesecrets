package secret

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/nacl/secretbox"
)

var _ Encryptor = &SecretboxEncryptor{}

type SecretboxEncryptor struct {
	logger *slog.Logger
}

func NewSecretboxEncryptor(logger *slog.Logger) *SecretboxEncryptor {
	return &SecretboxEncryptor{logger: logger}
}

func (e *SecretboxEncryptor) Encrypt(ctx context.Context, passphrase, message string) ([]byte, error) {
	key, random, err := e.makeKey(ctx, passphrase)
	if err != nil {
		return nil, err
	}

	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		e.logger.LogAttrs(ctx, slog.LevelError, "Failed to generate nonce", slog.String("error", err.Error()))

		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	var out []byte
	out = append(out, random...)
	out = append(out, nonce[:]...)

	data := secretbox.Seal(out, []byte(message), &nonce, &key)
	e.logger.DebugContext(ctx, "Message encrypted")

	return data, nil
}

func (e *SecretboxEncryptor) makeKey(ctx context.Context, passphrase string) ([32]byte, []byte, error) {
	var key [32]byte
	copy(key[:], []byte(passphrase))

	var random []byte
	n := len(key) - len(passphrase)
	if n > 0 {
		random = make([]byte, n)
		if _, err := rand.Read(random); err != nil {
			e.logger.LogAttrs(ctx, slog.LevelError, "Failed to supplement passphrase with random bytes", slog.String("error", err.Error()))

			return key, nil, fmt.Errorf("supplement passphrase: %w", err)
		}
		copy(key[len(passphrase):], random)
	}

	return key, random, nil
}

func (e *SecretboxEncryptor) Decrypt(ctx context.Context, passphrase string, data []byte) (string, error) {
	key, nonce, box := e.splitData(passphrase, data)

	message, ok := secretbox.Open(nil, box, &nonce, &key)
	if !ok {
		e.logger.DebugContext(ctx, "Invalid passphrase")

		return "", ErrInvalidPassphrase
	}
	e.logger.DebugContext(ctx, "Message decrypted")

	return string(message), nil
}

func (e *SecretboxEncryptor) splitData(passphrase string, data []byte) ([32]byte, [24]byte, []byte) {
	var key [32]byte
	copy(key[:], passphrase)

	if len(key)-len(passphrase) > 0 {
		n := copy(key[len(passphrase):], data)
		data = data[n:]
	}

	var nonce [24]byte
	n := copy(nonce[:], data)

	box := data[n:]

	return key, nonce, box
}
