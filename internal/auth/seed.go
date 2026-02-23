package auth

import (
	"context"
	"fmt"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
)

// AdminCreator is the interface for creating a user in the database.
type AdminCreator interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

// SeedAdmin creates an admin user with the given email and password.
func SeedAdmin(ctx context.Context, users AdminCreator, email, password string) (db.User, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return db.User{}, fmt.Errorf("hashing admin password: %w", err)
	}

	user, err := users.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Name:         "Admin",
		Role:         "admin",
	})
	if err != nil {
		return db.User{}, fmt.Errorf("creating admin user: %w", err)
	}

	return user, nil
}
