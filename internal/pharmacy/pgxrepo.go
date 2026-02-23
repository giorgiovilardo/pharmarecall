package pharmacy

import (
	"context"
	"errors"
	"fmt"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxRepository implements all pharmacy port interfaces using pgx/sqlc.
type PgxRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPgxRepository creates a new PgxRepository.
func NewPgxRepository(pool *pgxpool.Pool, queries *db.Queries) *PgxRepository {
	return &PgxRepository{pool: pool, queries: queries}
}

func (r *PgxRepository) CreateWithOwner(ctx context.Context, p CreateParams, ownerPasswordHash string) (Pharmacy, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Pharmacy{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	row, err := qtx.CreatePharmacy(ctx, db.CreatePharmacyParams{
		Name:    p.Name,
		Address: p.Address,
		Phone:   p.Phone,
		Email:   p.Email,
	})
	if err != nil {
		return Pharmacy{}, fmt.Errorf("creating pharmacy: %w", err)
	}

	if _, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:        p.OwnerEmail,
		PasswordHash: ownerPasswordHash,
		Name:         p.OwnerName,
		Role:         "owner",
		PharmacyID:   pgtype.Int8{Int64: row.ID, Valid: true},
	}); err != nil {
		return Pharmacy{}, mapDuplicateEmail(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Pharmacy{}, fmt.Errorf("committing transaction: %w", err)
	}

	return Pharmacy{
		ID:      row.ID,
		Name:    row.Name,
		Address: row.Address,
		Phone:   row.Phone,
		Email:   row.Email,
	}, nil
}

func (r *PgxRepository) GetByID(ctx context.Context, id int64) (Pharmacy, error) {
	row, err := r.queries.GetPharmacyByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Pharmacy{}, ErrNotFound
		}
		return Pharmacy{}, fmt.Errorf("querying pharmacy by id: %w", err)
	}
	return Pharmacy{
		ID:      row.ID,
		Name:    row.Name,
		Address: row.Address,
		Phone:   row.Phone,
		Email:   row.Email,
	}, nil
}

func (r *PgxRepository) List(ctx context.Context) ([]Summary, error) {
	rows, err := r.queries.ListPharmacies(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing pharmacies: %w", err)
	}
	summaries := make([]Summary, len(rows))
	for i, row := range rows {
		summaries[i] = Summary{
			ID:             row.ID,
			Name:           row.Name,
			Address:        row.Address,
			PersonnelCount: row.PersonnelCount,
		}
	}
	return summaries, nil
}

func (r *PgxRepository) Update(ctx context.Context, p UpdateParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).UpdatePharmacy(ctx, db.UpdatePharmacyParams{
		ID:      p.ID,
		Name:    p.Name,
		Address: p.Address,
		Phone:   p.Phone,
		Email:   p.Email,
	}); err != nil {
		return fmt.Errorf("updating pharmacy: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) ListPersonnel(ctx context.Context, pharmacyID int64) ([]PersonnelMember, error) {
	rows, err := r.queries.ListUsersByPharmacy(ctx, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("listing personnel: %w", err)
	}
	members := make([]PersonnelMember, len(rows))
	for i, row := range rows {
		members[i] = PersonnelMember{
			ID:    row.ID,
			Name:  row.Name,
			Email: row.Email,
			Role:  row.Role,
		}
	}
	return members, nil
}

func (r *PgxRepository) CreatePersonnel(ctx context.Context, p CreatePersonnelParams, passwordHash string) (PersonnelMember, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return PersonnelMember{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := r.queries.WithTx(tx).CreateUser(ctx, db.CreateUserParams{
		Email:        p.Email,
		PasswordHash: passwordHash,
		Name:         p.Name,
		Role:         p.Role,
		PharmacyID:   pgtype.Int8{Int64: p.PharmacyID, Valid: true},
	})
	if err != nil {
		return PersonnelMember{}, mapDuplicateEmail(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return PersonnelMember{}, fmt.Errorf("committing transaction: %w", err)
	}

	return PersonnelMember{
		ID:    row.ID,
		Name:  row.Name,
		Email: row.Email,
		Role:  row.Role,
	}, nil
}

func mapDuplicateEmail(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicateEmail
	}
	return err
}
