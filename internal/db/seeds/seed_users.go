package seeds

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

type userDTO struct {
	Email    string
	Password string
}

func createUser(ctx context.Context, db *sqlc.Queries, dto userDTO) (*sqlc.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error generating password hash: %w", err)
	}

	user, err := db.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        dto.Email,
		PasswordHash: string(hash),
	})
	return &user, err
}

func RunUsersSeeds(db *sqlc.Queries) {
	ctx := context.Background()

	users := []userDTO{
		{"user1@example.com", "password"},
		{"user2@example.com", "password"},
		{"user3@example.com", "password"},
	}

	for _, u := range users {
		user, err := createUser(ctx, db, u)
		if err != nil {
			panic(fmt.Errorf("error seeding user: %w", err))
		}
		slog.Info("seed user", "email", user.Email)
	}
}
