package seeds

import (
	"context"
	"log"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

func RunUsersSeeds(db *sqlc.Queries) {
	ctx := context.Background()

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		panic("error generating password hash")
	}

	_, err = db.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        "user@example.com",
		PasswordHash: string(hash),
	})
	if err != nil {
		log.Printf("error seeding user: %v", err)
	}

}
