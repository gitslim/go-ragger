package main

import (
	"context"
	"flag"
	"log"

	"github.com/gitslim/go-ragger/internal/db/seeds"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := flag.String("dsn", "", "DSN string")
	flag.Parse()

	pool, err := pgxpool.New(context.Background(), *dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	db := sqlc.New(pool)

	seeds.RunUsersSeeds(db)
}
