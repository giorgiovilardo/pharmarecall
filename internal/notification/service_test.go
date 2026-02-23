package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/giorgiovilardo/pharmarecall/internal/notification"
)

// --- Mocks ---

type mockCreator struct {
	called  bool
	params  []mockCreateCall
	err     error
}

type mockCreateCall struct {
	pharmacyID     int64
	prescriptionID int64
	transitionType string
}

func (m *mockCreator) Create(_ context.Context, pharmacyID, prescriptionID int64, transitionType string) error {
	m.called = true
	m.params = append(m.params, mockCreateCall{pharmacyID, prescriptionID, transitionType})
	return m.err
}

type mockLister struct {
	result []notification.Notification
	err    error
}

func (m *mockLister) ListByPharmacy(_ context.Context, _ int64) ([]notification.Notification, error) {
	return m.result, m.err
}

type mockReader struct {
	called     bool
	id         int64
	pharmacyID int64
	err        error
}

func (m *mockReader) MarkRead(_ context.Context, id, pharmacyID int64) error {
	m.called = true
	m.id = id
	m.pharmacyID = pharmacyID
	return m.err
}

type mockAllReader struct {
	called     bool
	pharmacyID int64
	err        error
}

func (m *mockAllReader) MarkAllRead(_ context.Context, pharmacyID int64) error {
	m.called = true
	m.pharmacyID = pharmacyID
	return m.err
}

type mockCounter struct {
	result int64
	err    error
}

func (m *mockCounter) CountUnread(_ context.Context, _ int64) (int64, error) {
	return m.result, m.err
}

// --- GenerateApproaching tests ---

func TestGenerateApproachingCreatesSingleNotification(t *testing.T) {
	creator := &mockCreator{}
	svc := notification.NewServiceWith(notification.ServiceDeps{Creator: creator})

	err := svc.GenerateApproaching(context.Background(), 7, []int64{42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !creator.called {
		t.Fatal("expected creator to be called")
	}
	if len(creator.params) != 1 {
		t.Fatalf("expected 1 call, got %d", len(creator.params))
	}
	p := creator.params[0]
	if p.pharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", p.pharmacyID)
	}
	if p.prescriptionID != 42 {
		t.Errorf("prescriptionID = %d, want 42", p.prescriptionID)
	}
	if p.transitionType != notification.TransitionApproaching {
		t.Errorf("transitionType = %q, want %q", p.transitionType, notification.TransitionApproaching)
	}
}

func TestGenerateApproachingSkipsEmpty(t *testing.T) {
	creator := &mockCreator{}
	svc := notification.NewServiceWith(notification.ServiceDeps{Creator: creator})

	err := svc.GenerateApproaching(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creator.called {
		t.Fatal("expected creator not to be called for empty slice")
	}
}

func TestGenerateApproachingMultiple(t *testing.T) {
	creator := &mockCreator{}
	svc := notification.NewServiceWith(notification.ServiceDeps{Creator: creator})

	err := svc.GenerateApproaching(context.Background(), 7, []int64{10, 20, 30})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(creator.params) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(creator.params))
	}
	for i, wantID := range []int64{10, 20, 30} {
		if creator.params[i].prescriptionID != wantID {
			t.Errorf("call %d: prescriptionID = %d, want %d", i, creator.params[i].prescriptionID, wantID)
		}
	}
}

func TestGenerateApproachingPropagatesError(t *testing.T) {
	creator := &mockCreator{err: errors.New("db error")}
	svc := notification.NewServiceWith(notification.ServiceDeps{Creator: creator})

	err := svc.GenerateApproaching(context.Background(), 7, []int64{42})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- List tests ---

func TestListReturnsNotifications(t *testing.T) {
	notifs := []notification.Notification{
		{ID: 1, PharmacyID: 7, MedicationName: "Tachipirina"},
		{ID: 2, PharmacyID: 7, MedicationName: "Aspirina"},
	}
	lister := &mockLister{result: notifs}
	svc := notification.NewServiceWith(notification.ServiceDeps{Lister: lister})

	got, err := svc.List(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestListPropagatesError(t *testing.T) {
	lister := &mockLister{err: errors.New("db error")}
	svc := notification.NewServiceWith(notification.ServiceDeps{Lister: lister})

	_, err := svc.List(context.Background(), 7)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- MarkRead tests ---

func TestMarkReadCallsReader(t *testing.T) {
	reader := &mockReader{}
	svc := notification.NewServiceWith(notification.ServiceDeps{Reader: reader})

	err := svc.MarkRead(context.Background(), 42, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reader.called {
		t.Fatal("expected reader to be called")
	}
	if reader.id != 42 {
		t.Errorf("id = %d, want 42", reader.id)
	}
	if reader.pharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", reader.pharmacyID)
	}
}

func TestMarkReadPropagatesError(t *testing.T) {
	reader := &mockReader{err: errors.New("db error")}
	svc := notification.NewServiceWith(notification.ServiceDeps{Reader: reader})

	err := svc.MarkRead(context.Background(), 42, 7)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- MarkAllRead tests ---

func TestMarkAllReadCallsAllReader(t *testing.T) {
	allReader := &mockAllReader{}
	svc := notification.NewServiceWith(notification.ServiceDeps{AllReader: allReader})

	err := svc.MarkAllRead(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allReader.called {
		t.Fatal("expected allReader to be called")
	}
	if allReader.pharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", allReader.pharmacyID)
	}
}

// --- CountUnread tests ---

func TestCountUnreadReturnsCount(t *testing.T) {
	counter := &mockCounter{result: 5}
	svc := notification.NewServiceWith(notification.ServiceDeps{Counter: counter})

	got, err := svc.CountUnread(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 5 {
		t.Errorf("count = %d, want 5", got)
	}
}

func TestCountUnreadPropagatesError(t *testing.T) {
	counter := &mockCounter{err: errors.New("db error")}
	svc := notification.NewServiceWith(notification.ServiceDeps{Counter: counter})

	_, err := svc.CountUnread(context.Background(), 7)
	if err == nil {
		t.Fatal("expected error")
	}
}
