package patient

import (
	"context"
	"errors"
	"fmt"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxRepository implements all patient port interfaces using pgx/sqlc.
type PgxRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPgxRepository creates a new PgxRepository.
func NewPgxRepository(pool *pgxpool.Pool, queries *db.Queries) *PgxRepository {
	return &PgxRepository{pool: pool, queries: queries}
}

func (r *PgxRepository) Create(ctx context.Context, p CreateParams) (Patient, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Patient{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := r.queries.WithTx(tx).CreatePatient(ctx, db.CreatePatientParams{
		PharmacyID:      p.PharmacyID,
		FirstName:       p.FirstName,
		LastName:        p.LastName,
		Phone:           p.Phone,
		Email:           p.Email,
		DeliveryAddress: p.DeliveryAddress,
		Fulfillment:     p.Fulfillment,
		Notes:           p.Notes,
	})
	if err != nil {
		return Patient{}, fmt.Errorf("creating patient: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Patient{}, fmt.Errorf("committing transaction: %w", err)
	}

	return mapPatient(row), nil
}

func (r *PgxRepository) GetByID(ctx context.Context, id int64) (Patient, error) {
	row, err := r.queries.GetPatientByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Patient{}, ErrNotFound
		}
		return Patient{}, fmt.Errorf("querying patient by id: %w", err)
	}
	return mapPatient(row), nil
}

func (r *PgxRepository) List(ctx context.Context, pharmacyID int64) ([]Summary, error) {
	rows, err := r.queries.ListPatientsByPharmacy(ctx, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("listing patients: %w", err)
	}
	summaries := make([]Summary, len(rows))
	for i, row := range rows {
		summaries[i] = Summary{
			ID:        row.ID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Phone:     row.Phone,
			Email:     row.Email,
			Consensus: row.Consensus,
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

	if err := r.queries.WithTx(tx).UpdatePatient(ctx, db.UpdatePatientParams{
		ID:              p.ID,
		FirstName:       p.FirstName,
		LastName:        p.LastName,
		Phone:           p.Phone,
		Email:           p.Email,
		DeliveryAddress: p.DeliveryAddress,
		Fulfillment:     p.Fulfillment,
		Notes:           p.Notes,
	}); err != nil {
		return fmt.Errorf("updating patient: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) SetConsensus(ctx context.Context, id int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).SetPatientConsensus(ctx, id); err != nil {
		return fmt.Errorf("setting consensus: %w", err)
	}

	return tx.Commit(ctx)
}

func mapPatient(row db.Patient) Patient {
	p := Patient{
		ID:              row.ID,
		PharmacyID:      row.PharmacyID,
		FirstName:       row.FirstName,
		LastName:        row.LastName,
		Phone:           row.Phone,
		Email:           row.Email,
		DeliveryAddress: row.DeliveryAddress,
		Fulfillment:     row.Fulfillment,
		Notes:           row.Notes,
		Consensus:       row.Consensus,
	}
	if row.ConsensusDate.Valid {
		d := row.ConsensusDate.Time.Format("2006-01-02")
		p.ConsensusDate = &d
	}
	return p
}
