package db

import (
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDb creates a new db instance
func NewDb(pool *pgxpool.Pool) *sqlc.Queries {
	return sqlc.New(pool)
}
