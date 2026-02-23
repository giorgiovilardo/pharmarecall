package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/dbutil"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Ensure PgxRepository satisfies Repository at compile time.
var _ Repository = (*PgxRepository)(nil)

// PgxRepository implements all order port interfaces using pgx/sqlc.
type PgxRepository struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewPgxRepository creates a new PgxRepository.
func NewPgxRepository(pool *pgxpool.Pool, queries *db.Queries) *PgxRepository {
	return &PgxRepository{pool: pool, queries: queries}
}

func (r *PgxRepository) Create(ctx context.Context, p CreateParams) (Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Order{}, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	row, err := r.queries.WithTx(tx).CreateOrder(ctx, db.CreateOrderParams{
		PrescriptionID:         p.PrescriptionID,
		CycleStartDate:         dbutil.TimeToDate(p.CycleStartDate),
		EstimatedDepletionDate: dbutil.TimeToDate(p.EstimatedDepletionDate),
		Status:                 StatusPending,
	})
	if err != nil {
		return Order{}, fmt.Errorf("creating order: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("committing transaction: %w", err)
	}

	return mapOrder(row), nil
}

func (r *PgxRepository) HasActiveOrder(ctx context.Context, prescriptionID int64, cycleStartDate time.Time) (bool, error) {
	_, err := r.queries.GetActiveOrderByPrescription(ctx, db.GetActiveOrderByPrescriptionParams{
		PrescriptionID: prescriptionID,
		CycleStartDate: dbutil.TimeToDate(cycleStartDate),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("checking active order: %w", err)
	}
	return true, nil
}

func (r *PgxRepository) ListDashboard(ctx context.Context, pharmacyID int64) ([]DashboardEntry, error) {
	rows, err := r.queries.ListDashboardOrders(ctx, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("listing dashboard orders: %w", err)
	}
	result := make([]DashboardEntry, len(rows))
	for i, row := range rows {
		result[i] = DashboardEntry{
			OrderID:                row.OrderID,
			PrescriptionID:         row.PrescriptionID,
			CycleStartDate:         row.CycleStartDate.Time,
			EstimatedDepletionDate: row.EstimatedDepletionDate.Time,
			OrderStatus:            row.OrderStatus,
			MedicationName:         row.MedicationName,
			UnitsPerBox:            int(row.UnitsPerBox),
			DailyConsumption:       dbutil.NumericToFloat64(row.DailyConsumption),
			BoxStartDate:           row.BoxStartDate.Time,
			PatientID:              row.PatientID,
			FirstName:              row.FirstName,
			LastName:               row.LastName,
			Fulfillment:            row.Fulfillment,
			DeliveryAddress:        row.DeliveryAddress,
			Phone:                  row.Phone,
			Email:                  row.Email,
		}
	}
	return result, nil
}

func (r *PgxRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.queries.WithTx(tx).UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     id,
		Status: status,
	}); err != nil {
		return fmt.Errorf("updating order status: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PgxRepository) GetByID(ctx context.Context, id int64) (Order, error) {
	row, err := r.queries.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Order{}, ErrNotFound
		}
		return Order{}, fmt.Errorf("querying order by id: %w", err)
	}
	return mapOrder(row), nil
}

func (r *PgxRepository) ListPrescriptionsForPharmacy(ctx context.Context, pharmacyID int64) ([]PrescriptionSummary, error) {
	rows, err := r.queries.ListPrescriptionsInLookahead(ctx, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("listing prescriptions for lookahead: %w", err)
	}
	result := make([]PrescriptionSummary, len(rows))
	for i, row := range rows {
		result[i] = PrescriptionSummary{
			ID:               row.PrescriptionID,
			PatientID:        row.PatientID,
			UnitsPerBox:      int(row.UnitsPerBox),
			DailyConsumption: dbutil.NumericToFloat64(row.DailyConsumption),
			BoxStartDate:     row.BoxStartDate.Time,
		}
	}
	return result, nil
}

func mapOrder(row db.Order) Order {
	return Order{
		ID:                     row.ID,
		PrescriptionID:         row.PrescriptionID,
		CycleStartDate:         row.CycleStartDate.Time,
		EstimatedDepletionDate: row.EstimatedDepletionDate.Time,
		Status:                 row.Status,
	}
}
