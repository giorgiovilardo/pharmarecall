package order

import (
	"context"
	"fmt"
	"time"
)

// ServiceDeps holds individual port interfaces — used by tests to inject only what's needed.
type ServiceDeps struct {
	Creator            OrderCreator
	ActiveChecker      ActiveOrderChecker
	Dashboard          DashboardLister
	StatusUpdater      OrderStatusUpdater
	Getter             OrderGetter
	PrescriptionLister PrescriptionLookaheadLister
	Refiller           PrescriptionRefiller
}

// Service contains order domain business logic.
type Service struct {
	deps ServiceDeps
}

// NewService is the production constructor — takes a Repository (satisfies all ports)
// and a PrescriptionRefiller for cross-domain refill on fulfillment.
func NewService(repo Repository, refiller PrescriptionRefiller) *Service {
	return &Service{deps: ServiceDeps{
		Creator:            repo,
		ActiveChecker:      repo,
		Dashboard:          repo,
		StatusUpdater:      repo,
		Getter:             repo,
		PrescriptionLister: repo,
		Refiller:           refiller,
	}}
}

// NewServiceWith is the test constructor — inject only what you need, rest stays nil.
func NewServiceWith(d ServiceDeps) *Service {
	return &Service{deps: d}
}

// EnsureOrders creates pending orders for prescriptions in the lookahead window
// that don't already have an active order for the current cycle.
func (s *Service) EnsureOrders(ctx context.Context, pharmacyID int64, now time.Time, lookaheadDays int) error {
	prescriptions, err := s.deps.PrescriptionLister.ListPrescriptionsForPharmacy(ctx, pharmacyID)
	if err != nil {
		return fmt.Errorf("listing prescriptions: %w", err)
	}

	for _, rx := range prescriptions {
		if rx.DaysRemaining(now) > lookaheadDays {
			continue
		}

		active, err := s.deps.ActiveChecker.HasActiveOrder(ctx, rx.ID, rx.BoxStartDate)
		if err != nil {
			return fmt.Errorf("checking active order for prescription %d: %w", rx.ID, err)
		}
		if active {
			continue
		}

		_, err = s.deps.Creator.Create(ctx, CreateParams{
			PrescriptionID:         rx.ID,
			CycleStartDate:         rx.BoxStartDate,
			EstimatedDepletionDate: rx.EstimatedDepletionDate(),
		})
		if err != nil {
			return fmt.Errorf("creating order for prescription %d: %w", rx.ID, err)
		}
	}

	return nil
}

// ListDashboard returns dashboard entries for a pharmacy.
func (s *Service) ListDashboard(ctx context.Context, pharmacyID int64) ([]DashboardEntry, error) {
	entries, err := s.deps.Dashboard.ListDashboard(ctx, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("listing dashboard: %w", err)
	}
	return entries, nil
}

// AdvanceStatus moves an order to the next status in the lifecycle.
// When transitioning to fulfilled, it also records a prescription refill.
func (s *Service) AdvanceStatus(ctx context.Context, orderID int64, now time.Time) error {
	o, err := s.deps.Getter.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("getting order: %w", err)
	}

	next := NextStatus(o.Status)
	if next == "" {
		return ErrInvalidTransition
	}

	if err := s.deps.StatusUpdater.UpdateStatus(ctx, orderID, next); err != nil {
		return fmt.Errorf("updating order status: %w", err)
	}

	if next == StatusFulfilled {
		if err := s.deps.Refiller.RecordRefill(ctx, o.PrescriptionID, now); err != nil {
			return fmt.Errorf("recording prescription refill: %w", err)
		}
	}

	return nil
}
