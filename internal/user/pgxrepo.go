package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure PgxRepository satisfies Repository at compile time.
var _ Repository = (*PgxRepository)(nil)

// PgxRepository implements all user port interfaces using pgx/sqlc.
type PgxRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPgxRepository creates a new PgxRepository.
func NewPgxRepository(pool *pgxpool.Pool, queries *db.Queries) *PgxRepository {
	return &PgxRepository{pool: pool, queries: queries}
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
		ID:         row.ID,
		Email:      row.Email,
		Name:       row.Name,
		Role:       row.Role,
		PharmacyID: row.PharmacyID.Int64,
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
		ID:         row.ID,
		Email:      row.Email,
		Name:       row.Name,
		Role:       row.Role,
		PharmacyID: row.PharmacyID.Int64,
	}, row.PasswordHash, nil
}

func (r *PgxRepository) UpdatePassword(ctx context.Context, id int64, hash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: hash,
	}); err != nil {
		return fmt.Errorf("updating user password: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) Create(ctx context.Context, email, passwordHash, name, role string) (User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := r.queries.WithTx(tx).CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         role,
	})
	if err != nil {
		return User{}, fmt.Errorf("creating user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("committing transaction: %w", err)
	}

	return User{
		ID:    row.ID,
		Email: row.Email,
		Name:  row.Name,
		Role:  row.Role,
	}, nil
}
