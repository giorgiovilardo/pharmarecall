package store

import (
	"context"
	"fmt"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreatePharmacyWithOwner returns a function that creates a pharmacy and its
// owner user inside a single transaction.
func CreatePharmacyWithOwner(pool *pgxpool.Pool, queries *db.Queries) web.CreatePharmacyWithOwnerFunc {
	return func(ctx context.Context, p db.CreatePharmacyParams, owner db.CreateUserParams) (db.Pharmacy, error) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return db.Pharmacy{}, fmt.Errorf("beginning transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		qtx := queries.WithTx(tx)

		pharmacy, err := qtx.CreatePharmacy(ctx, p)
		if err != nil {
			return db.Pharmacy{}, fmt.Errorf("creating pharmacy: %w", err)
		}

		owner.PharmacyID = pgtype.Int8{Int64: pharmacy.ID, Valid: true}
		if _, err := qtx.CreateUser(ctx, owner); err != nil {
			return db.Pharmacy{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return db.Pharmacy{}, fmt.Errorf("committing transaction: %w", err)
		}
		return pharmacy, nil
	}
}

// UpdatePharmacy returns a function that updates a pharmacy inside a transaction.
func UpdatePharmacy(pool *pgxpool.Pool, queries *db.Queries) web.UpdatePharmacyFunc {
	return func(ctx context.Context, arg db.UpdatePharmacyParams) error {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		if err := queries.WithTx(tx).UpdatePharmacy(ctx, arg); err != nil {
			return fmt.Errorf("updating pharmacy: %w", err)
		}
		return tx.Commit(ctx)
	}
}

// CreatePersonnel returns a function that creates a personnel user inside a transaction.
func CreatePersonnel(pool *pgxpool.Pool, queries *db.Queries) web.CreatePersonnelFunc {
	return func(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return db.User{}, fmt.Errorf("beginning transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		user, err := queries.WithTx(tx).CreateUser(ctx, arg)
		if err != nil {
			return db.User{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return db.User{}, fmt.Errorf("committing transaction: %w", err)
		}
		return user, nil
	}
}
