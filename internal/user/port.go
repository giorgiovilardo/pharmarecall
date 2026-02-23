package user

import "context"

// UserByEmailGetter fetches a user by email, returning their password hash for verification.
type UserByEmailGetter interface {
	GetByEmail(ctx context.Context, email string) (User, string, error)
}

// UserByIDGetter fetches a user by ID, returning their password hash for verification.
type UserByIDGetter interface {
	GetByID(ctx context.Context, id int64) (User, string, error)
}

// PasswordUpdater updates a user's password hash.
type PasswordUpdater interface {
	UpdatePassword(ctx context.Context, id int64, hash string) error
}

// UserCreator creates a new user.
type UserCreator interface {
	Create(ctx context.Context, email, passwordHash, name, role string) (User, error)
}

// Repository composes all ports — used only by NewService for convenient wiring.
type Repository interface {
	UserByEmailGetter
	UserByIDGetter
	PasswordUpdater
	UserCreator
}
