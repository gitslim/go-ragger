package db

import (
	"context"
	"fmt"
	"time"

	"github.com/gitslim/go-ragger/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

// NewPgxPool creates a new postgres pool.
func NewPgxPool(cfg *config.ServerConfig) (*pgxpool.Pool, error) {
	c, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	c.MaxConns = 20
	c.MinConns = 5
	c.MaxConnLifetime = time.Hour
	c.MaxConnIdleTime = 30 * time.Minute
	c.HealthCheckPeriod = time.Minute

	ctx := context.Background()

	pool, err := pgxpool.NewWithConfig(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("create pool failed: %w", err)
	}

	return pool, nil
}

// RegisterDBPoolHooks starts and stops the DB pool.
func RegisterDBPoolHooks(lc fx.Lifecycle, pool *pgxpool.Pool) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return fmt.Errorf("pgxpool ping failed: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
}
