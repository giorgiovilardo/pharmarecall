package order_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/order"
)

func date(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// --- Mocks ---

type mockPrescriptionLister struct {
	result []order.PrescriptionSummary
	err    error
}

func (m *mockPrescriptionLister) ListPrescriptionsForPharmacy(_ context.Context, _ int64) ([]order.PrescriptionSummary, error) {
	return m.result, m.err
}

type mockActiveChecker struct {
	active bool
	err    error
}

func (m *mockActiveChecker) HasActiveOrder(_ context.Context, _ int64, _ time.Time) (bool, error) {
	return m.active, m.err
}

type mockCreator struct {
	called  bool
	params  []order.CreateParams
	counter int64
	err     error
}

func (m *mockCreator) Create(_ context.Context, p order.CreateParams) (order.Order, error) {
	m.called = true
	m.params = append(m.params, p)
	m.counter++
	return order.Order{
		ID:                     m.counter,
		PrescriptionID:         p.PrescriptionID,
		CycleStartDate:         p.CycleStartDate,
		EstimatedDepletionDate: p.EstimatedDepletionDate,
		Status:                 order.StatusPending,
	}, m.err
}

// --- EnsureOrders tests ---

func TestEnsureOrdersCreatesOrderForPrescriptionInWindow(t *testing.T) {
	// Prescription: 30 units at 1/day, started Jan 1 → depletes Jan 31.
	// Now is Jan 27 → 4 days remaining → within 7-day window.
	lister := &mockPrescriptionLister{result: []order.PrescriptionSummary{
		{ID: 1, PatientID: 10, UnitsPerBox: 30, DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
	}}
	checker := &mockActiveChecker{active: false}
	creator := &mockCreator{}

	svc := order.NewServiceWith(order.ServiceDeps{
		PrescriptionLister: lister,
		ActiveChecker:      checker,
		Creator:            creator,
	})

	err := svc.EnsureOrders(context.Background(), 1, date(2026, 1, 27), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !creator.called {
		t.Fatal("expected order to be created")
	}
	if len(creator.params) != 1 {
		t.Fatalf("expected 1 order created, got %d", len(creator.params))
	}
	p := creator.params[0]
	if p.PrescriptionID != 1 {
		t.Errorf("PrescriptionID = %d, want 1", p.PrescriptionID)
	}
	if !p.CycleStartDate.Equal(date(2026, 1, 1)) {
		t.Errorf("CycleStartDate = %s, want 2026-01-01", p.CycleStartDate.Format("2006-01-02"))
	}
	if !p.EstimatedDepletionDate.Equal(date(2026, 1, 31)) {
		t.Errorf("EstimatedDepletionDate = %s, want 2026-01-31", p.EstimatedDepletionDate.Format("2006-01-02"))
	}
}

func TestEnsureOrdersSkipsPrescriptionOutsideWindow(t *testing.T) {
	// Prescription: 30 units at 1/day, started Jan 1 → depletes Jan 31.
	// Now is Jan 10 → 21 days remaining → outside 7-day window.
	lister := &mockPrescriptionLister{result: []order.PrescriptionSummary{
		{ID: 1, PatientID: 10, UnitsPerBox: 30, DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
	}}
	checker := &mockActiveChecker{active: false}
	creator := &mockCreator{}

	svc := order.NewServiceWith(order.ServiceDeps{
		PrescriptionLister: lister,
		ActiveChecker:      checker,
		Creator:            creator,
	})

	err := svc.EnsureOrders(context.Background(), 1, date(2026, 1, 10), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creator.called {
		t.Fatal("expected no order to be created for prescription outside window")
	}
}

func TestEnsureOrdersSkipsWhenActiveOrderExists(t *testing.T) {
	// Prescription in window but already has an active order.
	lister := &mockPrescriptionLister{result: []order.PrescriptionSummary{
		{ID: 1, PatientID: 10, UnitsPerBox: 30, DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
	}}
	checker := &mockActiveChecker{active: true}
	creator := &mockCreator{}

	svc := order.NewServiceWith(order.ServiceDeps{
		PrescriptionLister: lister,
		ActiveChecker:      checker,
		Creator:            creator,
	})

	err := svc.EnsureOrders(context.Background(), 1, date(2026, 1, 27), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creator.called {
		t.Fatal("expected no order to be created when active order exists")
	}
}

func TestEnsureOrdersCreatesOrderForDepletedPrescription(t *testing.T) {
	// Prescription: 30 units at 1/day, started Jan 1 → depleted Jan 31.
	// Now is Feb 5 → -5 days remaining → depleted, still needs order.
	lister := &mockPrescriptionLister{result: []order.PrescriptionSummary{
		{ID: 1, PatientID: 10, UnitsPerBox: 30, DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
	}}
	checker := &mockActiveChecker{active: false}
	creator := &mockCreator{}

	svc := order.NewServiceWith(order.ServiceDeps{
		PrescriptionLister: lister,
		ActiveChecker:      checker,
		Creator:            creator,
	})

	err := svc.EnsureOrders(context.Background(), 1, date(2026, 2, 5), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !creator.called {
		t.Fatal("expected order to be created for depleted prescription")
	}
}

func TestEnsureOrdersNoPrescriptions(t *testing.T) {
	lister := &mockPrescriptionLister{result: nil}
	svc := order.NewServiceWith(order.ServiceDeps{PrescriptionLister: lister})

	err := svc.EnsureOrders(context.Background(), 1, date(2026, 1, 27), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- ListDashboard tests ---

type mockDashboardLister struct {
	result []order.DashboardEntry
	err    error
}

func (m *mockDashboardLister) ListDashboard(_ context.Context, _ int64) ([]order.DashboardEntry, error) {
	return m.result, m.err
}

func TestListDashboardSuccess(t *testing.T) {
	entries := []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", OrderStatus: order.StatusPending},
		{OrderID: 2, MedicationName: "Aspirina", OrderStatus: order.StatusPrepared},
	}
	dashboard := &mockDashboardLister{result: entries}
	svc := order.NewServiceWith(order.ServiceDeps{Dashboard: dashboard})

	got, err := svc.ListDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

// --- AdvanceStatus tests ---

type mockGetter struct {
	result order.Order
	err    error
}

func (m *mockGetter) GetByID(_ context.Context, _ int64) (order.Order, error) {
	return m.result, m.err
}

type mockStatusUpdater struct {
	called    bool
	id        int64
	newStatus string
	err       error
}

func (m *mockStatusUpdater) UpdateStatus(_ context.Context, id int64, status string) error {
	m.called = true
	m.id = id
	m.newStatus = status
	return m.err
}

type mockRefiller struct {
	called         bool
	prescriptionID int64
	newStartDate   time.Time
	err            error
}

func (m *mockRefiller) RecordRefill(_ context.Context, prescriptionID int64, newStartDate time.Time) error {
	m.called = true
	m.prescriptionID = prescriptionID
	m.newStartDate = newStartDate
	return m.err
}

func TestAdvanceStatusPendingToPrepared(t *testing.T) {
	getter := &mockGetter{result: order.Order{ID: 1, Status: order.StatusPending}}
	updater := &mockStatusUpdater{}
	refiller := &mockRefiller{}
	svc := order.NewServiceWith(order.ServiceDeps{Getter: getter, StatusUpdater: updater, Refiller: refiller})

	err := svc.AdvanceStatus(context.Background(), 1, date(2026, 2, 23))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updater.called {
		t.Fatal("UpdateStatus was not called")
	}
	if updater.newStatus != order.StatusPrepared {
		t.Errorf("newStatus = %q, want %q", updater.newStatus, order.StatusPrepared)
	}
	if refiller.called {
		t.Error("RecordRefill should not be called for pending->prepared")
	}
}

func TestAdvanceStatusPreparedToFulfilled(t *testing.T) {
	now := date(2026, 2, 23)
	getter := &mockGetter{result: order.Order{ID: 1, PrescriptionID: 42, Status: order.StatusPrepared}}
	updater := &mockStatusUpdater{}
	refiller := &mockRefiller{}
	svc := order.NewServiceWith(order.ServiceDeps{Getter: getter, StatusUpdater: updater, Refiller: refiller})

	err := svc.AdvanceStatus(context.Background(), 1, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updater.newStatus != order.StatusFulfilled {
		t.Errorf("newStatus = %q, want %q", updater.newStatus, order.StatusFulfilled)
	}
	if !refiller.called {
		t.Fatal("RecordRefill was not called on fulfillment")
	}
	if refiller.prescriptionID != 42 {
		t.Errorf("prescriptionID = %d, want 42", refiller.prescriptionID)
	}
	if !refiller.newStartDate.Equal(now) {
		t.Errorf("newStartDate = %s, want %s", refiller.newStartDate.Format("2006-01-02"), now.Format("2006-01-02"))
	}
}

func TestAdvanceStatusRefillErrorPropagates(t *testing.T) {
	getter := &mockGetter{result: order.Order{ID: 1, PrescriptionID: 42, Status: order.StatusPrepared}}
	updater := &mockStatusUpdater{}
	refiller := &mockRefiller{err: errors.New("refill failed")}
	svc := order.NewServiceWith(order.ServiceDeps{Getter: getter, StatusUpdater: updater, Refiller: refiller})

	err := svc.AdvanceStatus(context.Background(), 1, date(2026, 2, 23))
	if err == nil {
		t.Fatal("expected error when refill fails")
	}
}

func TestAdvanceStatusFulfilledIsTerminal(t *testing.T) {
	getter := &mockGetter{result: order.Order{ID: 1, Status: order.StatusFulfilled}}
	updater := &mockStatusUpdater{}
	svc := order.NewServiceWith(order.ServiceDeps{Getter: getter, StatusUpdater: updater})

	err := svc.AdvanceStatus(context.Background(), 1, date(2026, 2, 23))
	if err == nil {
		t.Fatal("expected error for terminal status")
	}
	if !errors.Is(err, order.ErrInvalidTransition) {
		t.Errorf("error = %v, want ErrInvalidTransition", err)
	}
	if updater.called {
		t.Error("UpdateStatus should not have been called")
	}
}

func TestAdvanceStatusNotFound(t *testing.T) {
	getter := &mockGetter{err: order.ErrNotFound}
	svc := order.NewServiceWith(order.ServiceDeps{Getter: getter})

	err := svc.AdvanceStatus(context.Background(), 999, date(2026, 2, 23))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, order.ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
}
