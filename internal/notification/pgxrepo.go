package notification

import (
	"context"
	"fmt"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/dbutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure PgxRepository satisfies Repository at compile time.
var _ Repository = (*PgxRepository)(nil)

// PgxRepository implements all notification port interfaces using pgx/sqlc.
type PgxRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPgxRepository creates a new PgxRepository.
func NewPgxRepository(pool *pgxpool.Pool, queries *db.Queries) *PgxRepository {
	return &PgxRepository{pool: pool, queries: queries}
}

func (r *PgxRepository) Create(ctx context.Context, pharmacyID, prescriptionID int64, transitionType string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).CreateNotification(ctx, db.CreateNotificationParams{
		PharmacyID:     pharmacyID,
		PrescriptionID: prescriptionID,
		TransitionType: transitionType,
	}); err != nil {
		return fmt.Errorf("creating notification: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) ListByPharmacy(ctx context.Context, pharmacyID int64) ([]Notification, error) {
	rows, err := r.queries.ListNotificationsByPharmacy(ctx, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("listing notifications: %w", err)
	}
	result := make([]Notification, len(rows))
	for i, row := range rows {
		result[i] = Notification{
			ID:               row.ID,
			PharmacyID:       row.PharmacyID,
			PrescriptionID:   row.PrescriptionID,
			TransitionType:   row.TransitionType,
			Read:             row.Read,
			CreatedAt:        row.CreatedAt.Time,
			MedicationName:   row.MedicationName,
			UnitsPerBox:      int(row.UnitsPerBox),
			DailyConsumption: dbutil.NumericToFloat64(row.DailyConsumption),
			BoxStartDate:     row.BoxStartDate.Time,
			PatientID:        row.PatientID,
			FirstName:        row.FirstName,
			LastName:         row.LastName,
		}
	}
	return result, nil
}

func (r *PgxRepository) MarkRead(ctx context.Context, id, pharmacyID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).MarkNotificationRead(ctx, db.MarkNotificationReadParams{
		ID:         id,
		PharmacyID: pharmacyID,
	}); err != nil {
		return fmt.Errorf("marking notification read: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) MarkAllRead(ctx context.Context, pharmacyID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).MarkAllNotificationsRead(ctx, pharmacyID); err != nil {
		return fmt.Errorf("marking all notifications read: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) CountUnread(ctx context.Context, pharmacyID int64) (int64, error) {
	count, err := r.queries.CountUnreadNotifications(ctx, pharmacyID)
	if err != nil {
		return 0, fmt.Errorf("counting unread notifications: %w", err)
	}
	return count, nil
}
