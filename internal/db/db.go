package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pg *pgxpool.Pool

func Close() {
	pg.Close()
}

func Init(ctx context.Context) error {
	if pg != nil {
		return nil
	}

	pool, err := pgxpool.New(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}

	pg = pool

	return nil
}

func Shutdown() {
	pg.Close()
}
