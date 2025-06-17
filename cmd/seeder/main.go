package main

import (
	"context"
	"flag"
	"log"

	"github.com/gitslim/go-ragger/internal/db/seeds"
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

	seeds.RunUsersSeeds(pool)
}
