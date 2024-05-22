package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ Store = &PgStore{}

type PgStore struct {
	pool *pgxpool.Pool
}

func NewPgStore(pool *pgxpool.Pool) *PgStore {
	return &PgStore{
		pool: pool,
	}
}

func (p *PgStore) Init(ctx context.Context) error {
	sql := `
		CREATE TABLE IF NOT EXISTS secrets (
			key      CHAR(255)   PRIMARY KEY,
			data     BYTEA       NOT NULL,
			attempts SMALLINT    NOT NULL,
			expireAt TIMESTAMPTZ NOT NULL
		)
	`
	_, err := p.pool.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("initialize db: %w", err)
	}

	return nil
}

func (p *PgStore) Load(ctx context.Context, key string) (Secret, error) {
	secret := Secret{}
	row := p.pool.QueryRow(ctx, "SELECT data, attempts, expireAt FROM secrets WHERE key=$1", key)
	err := row.Scan(&secret.data, &secret.attempts, &secret.exp)
	if errors.Is(err, pgx.ErrNoRows) {
		return secret, ErrNotFound
	}
	if err != nil {
		return secret, fmt.Errorf("select query: %w", err)
	}

	return secret, nil
}

func (p *PgStore) Save(ctx context.Context, key string, secret Secret) error {
	sql := `
		INSERT INTO secrets (key, data, attempts, expireAt)
		VALUES (@key, @data, @attempts, @expireAt)
		ON CONFLICT (key)
		DO UPDATE SET data = EXCLUDED.data, attempts = EXCLUDED.attempts, expireAt = EXCLUDED.expireAt
	`
	_, err := p.pool.Exec(ctx, sql, pgx.NamedArgs{
		"key":      key,
		"data":     secret.data,
		"attempts": secret.attempts,
		"expireAt": secret.exp,
	})
	if err != nil {
		return fmt.Errorf("upsert query: %w", err)
	}

	return nil
}

func (p *PgStore) Remove(ctx context.Context, key string) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM secrets WHERE key=$1", key)
	if err != nil {
		return fmt.Errorf("delete by key query: %w", err)
	}

	return nil
}

func (p *PgStore) Cleanup(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM secrets WHERE expireAt < now()")
	if err != nil {
		return fmt.Errorf("delete expired query: %w", err)
	}

	return nil
}
