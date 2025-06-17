package seeds

import (
	"context"
	"log"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func RunUsersSeeds(pool *pgxpool.Pool) {
	ctx := context.Background()
	q := sqlc.New(pool)

	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		panic("error generating password hash")
	}

	_, err = q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        "admin@foo.com",
		PasswordHash: string(hash),
	})
	if err != nil {
		log.Printf("error seeding admin user: %v", err)
	}

}
