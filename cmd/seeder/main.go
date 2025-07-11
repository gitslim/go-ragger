package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/gitslim/go-ragger/internal/db/seeds"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		panic("DSN environment variable not set")
	}

	flag.Parse()

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	db := sqlc.New(pool)

	seeds.RunUsersSeeds(db)
}
