package prescription

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgxRepository implements all prescription port interfaces using pgx/sqlc.
type PgxRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPgxRepository creates a new PgxRepository.
func NewPgxRepository(pool *pgxpool.Pool, queries *db.Queries) *PgxRepository {
	return &PgxRepository{pool: pool, queries: queries}
}

func (r *PgxRepository) Create(ctx context.Context, p CreateParams) (Prescription, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Prescription{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := r.queries.WithTx(tx).CreatePrescription(ctx, db.CreatePrescriptionParams{
		PatientID:        p.PatientID,
		MedicationName:   p.MedicationName,
		UnitsPerBox:      int32(p.UnitsPerBox),
		DailyConsumption: float64ToNumeric(p.DailyConsumption),
		BoxStartDate:     timeToDate(p.BoxStartDate),
	})
	if err != nil {
		return Prescription{}, fmt.Errorf("creating prescription: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Prescription{}, fmt.Errorf("committing transaction: %w", err)
	}

	return mapPrescription(row), nil
}

func (r *PgxRepository) GetByID(ctx context.Context, id int64) (Prescription, error) {
	row, err := r.queries.GetPrescriptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Prescription{}, ErrNotFound
		}
		return Prescription{}, fmt.Errorf("querying prescription by id: %w", err)
	}
	return mapPrescription(row), nil
}

func (r *PgxRepository) ListByPatient(ctx context.Context, patientID int64) ([]Prescription, error) {
	rows, err := r.queries.ListPrescriptionsByPatient(ctx, patientID)
	if err != nil {
		return nil, fmt.Errorf("listing prescriptions: %w", err)
	}
	result := make([]Prescription, len(rows))
	for i, row := range rows {
		result[i] = mapPrescription(row)
	}
	return result, nil
}

func (r *PgxRepository) Update(ctx context.Context, p UpdateParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).UpdatePrescription(ctx, db.UpdatePrescriptionParams{
		ID:               p.ID,
		MedicationName:   p.MedicationName,
		UnitsPerBox:      int32(p.UnitsPerBox),
		DailyConsumption: float64ToNumeric(p.DailyConsumption),
		BoxStartDate:     timeToDate(p.BoxStartDate),
	}); err != nil {
		return fmt.Errorf("updating prescription: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) RecordRefill(ctx context.Context, p RefillParams) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	// Get the current prescription to record history.
	current, err := qtx.GetPrescriptionByID(ctx, p.PrescriptionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("getting prescription for refill: %w", err)
	}

	// Calculate the old box end date (depletion date).
	dailyConsumption := numericToFloat64(current.DailyConsumption)
	days := math.Floor(float64(current.UnitsPerBox) / dailyConsumption)
	oldEnd := current.BoxStartDate.Time.AddDate(0, 0, int(days))

	// Insert refill history for the previous cycle.
	if err := qtx.InsertRefillHistory(ctx, db.InsertRefillHistoryParams{
		PrescriptionID: p.PrescriptionID,
		BoxStartDate:   current.BoxStartDate,
		BoxEndDate:     timeToDate(oldEnd),
	}); err != nil {
		return fmt.Errorf("inserting refill history: %w", err)
	}

	// Update the box start date to the new refill date.
	if err := qtx.UpdatePrescription(ctx, db.UpdatePrescriptionParams{
		ID:               p.PrescriptionID,
		MedicationName:   current.MedicationName,
		UnitsPerBox:      current.UnitsPerBox,
		DailyConsumption: current.DailyConsumption,
		BoxStartDate:     timeToDate(p.NewStartDate),
	}); err != nil {
		return fmt.Errorf("updating prescription start date: %w", err)
	}

	return tx.Commit(ctx)
}

func mapPrescription(row db.Prescription) Prescription {
	return Prescription{
		ID:               row.ID,
		PatientID:        row.PatientID,
		MedicationName:   row.MedicationName,
		UnitsPerBox:      int(row.UnitsPerBox),
		DailyConsumption: numericToFloat64(row.DailyConsumption),
		BoxStartDate:     row.BoxStartDate.Time,
	}
}

func float64ToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	s := new(big.Float).SetFloat64(f).Text('f', -1)
	n.ScanScientific(s)
	return n
}

func numericToFloat64(n pgtype.Numeric) float64 {
	f, _ := n.Float64Value()
	return f.Float64
}

func timeToDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}
