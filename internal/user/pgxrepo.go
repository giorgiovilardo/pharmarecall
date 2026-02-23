package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5"
)

// PgxRepository implements all user port interfaces using pgx/sqlc.
type PgxRepository struct {
	queries *db.Queries
}

// NewPgxRepository creates a new PgxRepository.
func NewPgxRepository(queries *db.Queries) *PgxRepository {
	return &PgxRepository{queries: queries}
}

func (r *PgxRepository) GetByEmail(ctx context.Context, email string) (User, string, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, "", ErrNotFound
		}
		return User{}, "", fmt.Errorf("querying user by email: %w", err)
	}
	return User{
		ID:    row.ID,
		Email: row.Email,
		Name:  row.Name,
		Role:  row.Role,
	}, row.PasswordHash, nil
}

func (r *PgxRepository) GetByID(ctx context.Context, id int64) (User, string, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, "", ErrNotFound
		}
		return User{}, "", fmt.Errorf("querying user by id: %w", err)
	}
	return User{
		ID:    row.ID,
		Email: row.Email,
		Name:  row.Name,
		Role:  row.Role,
	}, row.PasswordHash, nil
}

func (r *PgxRepository) UpdatePassword(ctx context.Context, id int64, hash string) error {
	if err := r.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: hash,
	}); err != nil {
		return fmt.Errorf("updating user password: %w", err)
	}
	return nil
}

func (r *PgxRepository) Create(ctx context.Context, email, passwordHash, name, role string) (User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         role,
	})
	if err != nil {
		return User{}, fmt.Errorf("creating user: %w", err)
	}
	return User{
		ID:    row.ID,
		Email: row.Email,
		Name:  row.Name,
		Role:  row.Role,
	}, nil
}
