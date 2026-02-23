package prescription_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/giorgiovilardo/pharmarecall/internal/prescription"
)

// --- Mocks ---

type mockCreator struct {
	called bool
	params prescription.CreateParams
	result prescription.Prescription
	err    error
}

func (m *mockCreator) Create(_ context.Context, p prescription.CreateParams) (prescription.Prescription, error) {
	m.called = true
	m.params = p
	return m.result, m.err
}

type mockGetter struct {
	result prescription.Prescription
	err    error
}

func (m *mockGetter) GetByID(_ context.Context, _ int64) (prescription.Prescription, error) {
	return m.result, m.err
}

type mockLister struct {
	result []prescription.Prescription
	err    error
}

func (m *mockLister) ListByPatient(_ context.Context, _ int64) ([]prescription.Prescription, error) {
	return m.result, m.err
}

type mockUpdater struct {
	called bool
	params prescription.UpdateParams
	err    error
}

func (m *mockUpdater) Update(_ context.Context, p prescription.UpdateParams) error {
	m.called = true
	m.params = p
	return m.err
}

type mockRefillRecorder struct {
	called bool
	params prescription.RefillParams
	err    error
}

func (m *mockRefillRecorder) RecordRefill(_ context.Context, p prescription.RefillParams) error {
	m.called = true
	m.params = p
	return m.err
}

type mockConsensusChecker struct {
	consensus bool
	err       error
}

func (m *mockConsensusChecker) HasConsensus(_ context.Context, _ int64) (bool, error) {
	return m.consensus, m.err
}

// --- Create tests ---

func TestCreateSuccess(t *testing.T) {
	creator := &mockCreator{result: prescription.Prescription{ID: 1, MedicationName: "Tachipirina"}}
	checker := &mockConsensusChecker{consensus: true}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Creator: creator, Consensus: checker})

	got, err := svc.Create(context.Background(), prescription.CreateParams{
		PatientID:        10,
		MedicationName:   "Tachipirina",
		UnitsPerBox:      30,
		DailyConsumption: 1,
		BoxStartDate:     date(2026, 1, 1),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if !creator.called {
		t.Fatal("Create was not called")
	}
}

func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name   string
		params prescription.CreateParams
		errStr string
	}{
		{
			name:   "missing medication name",
			params: prescription.CreateParams{PatientID: 1, UnitsPerBox: 30, DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
			errStr: "farmaco",
		},
		{
			name:   "zero units per box",
			params: prescription.CreateParams{PatientID: 1, MedicationName: "X", UnitsPerBox: 0, DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
			errStr: "unità",
		},
		{
			name:   "zero daily consumption",
			params: prescription.CreateParams{PatientID: 1, MedicationName: "X", UnitsPerBox: 30, DailyConsumption: 0, BoxStartDate: date(2026, 1, 1)},
			errStr: "consumo",
		},
		{
			name:   "zero box start date",
			params: prescription.CreateParams{PatientID: 1, MedicationName: "X", UnitsPerBox: 30, DailyConsumption: 1},
			errStr: "data",
		},
		{
			name:   "consumption equals units (box lasts less than 1 day)",
			params: prescription.CreateParams{PatientID: 1, MedicationName: "X", UnitsPerBox: 30, DailyConsumption: 30, BoxStartDate: date(2026, 1, 1)},
			errStr: "consumo giornaliero deve essere inferiore",
		},
		{
			name:   "consumption exceeds units",
			params: prescription.CreateParams{PatientID: 1, MedicationName: "X", UnitsPerBox: 10, DailyConsumption: 50, BoxStartDate: date(2026, 1, 1)},
			errStr: "consumo giornaliero deve essere inferiore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &mockConsensusChecker{consensus: true}
			svc := prescription.NewServiceWith(prescription.ServiceDeps{Consensus: checker})
			_, err := svc.Create(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.errStr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.errStr)
			}
		})
	}
}

func TestCreateBlocksWithoutConsensus(t *testing.T) {
	creator := &mockCreator{}
	checker := &mockConsensusChecker{consensus: false}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Creator: creator, Consensus: checker})

	_, err := svc.Create(context.Background(), prescription.CreateParams{
		PatientID:        10,
		MedicationName:   "Tachipirina",
		UnitsPerBox:      30,
		DailyConsumption: 1,
		BoxStartDate:     date(2026, 1, 1),
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, prescription.ErrNoConsensus) {
		t.Errorf("error = %v, want ErrNoConsensus", err)
	}
	if creator.called {
		t.Error("Create should not have been called")
	}
}

func TestCreateRepoError(t *testing.T) {
	creator := &mockCreator{err: errors.New("db down")}
	checker := &mockConsensusChecker{consensus: true}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Creator: creator, Consensus: checker})

	_, err := svc.Create(context.Background(), prescription.CreateParams{
		PatientID:        10,
		MedicationName:   "Tachipirina",
		UnitsPerBox:      30,
		DailyConsumption: 1,
		BoxStartDate:     date(2026, 1, 1),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- Update tests ---

func TestUpdateSuccess(t *testing.T) {
	updater := &mockUpdater{}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Updater: updater})

	err := svc.Update(context.Background(), prescription.UpdateParams{
		ID:               1,
		MedicationName:   "Tachipirina",
		UnitsPerBox:      60,
		DailyConsumption: 2,
		BoxStartDate:     date(2026, 1, 1),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updater.called {
		t.Fatal("Update was not called")
	}
}

func TestUpdateValidation(t *testing.T) {
	tests := []struct {
		name   string
		params prescription.UpdateParams
		errStr string
	}{
		{
			name:   "missing medication name",
			params: prescription.UpdateParams{ID: 1, UnitsPerBox: 30, DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
			errStr: "farmaco",
		},
		{
			name:   "zero units",
			params: prescription.UpdateParams{ID: 1, MedicationName: "X", DailyConsumption: 1, BoxStartDate: date(2026, 1, 1)},
			errStr: "unità",
		},
		{
			name:   "zero consumption",
			params: prescription.UpdateParams{ID: 1, MedicationName: "X", UnitsPerBox: 30, BoxStartDate: date(2026, 1, 1)},
			errStr: "consumo",
		},
		{
			name:   "consumption exceeds units",
			params: prescription.UpdateParams{ID: 1, MedicationName: "X", UnitsPerBox: 10, DailyConsumption: 20, BoxStartDate: date(2026, 1, 1)},
			errStr: "consumo giornaliero deve essere inferiore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockUpdater{}
			svc := prescription.NewServiceWith(prescription.ServiceDeps{Updater: updater})
			err := svc.Update(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if updater.called {
				t.Error("Update should not have been called")
			}
		})
	}
}

// --- Refill tests ---

func TestRecordRefillSuccess(t *testing.T) {
	recorder := &mockRefillRecorder{}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Refill: recorder})

	err := svc.RecordRefill(context.Background(), 1, date(2026, 2, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !recorder.called {
		t.Fatal("RecordRefill was not called")
	}
	if recorder.params.PrescriptionID != 1 {
		t.Errorf("PrescriptionID = %d, want 1", recorder.params.PrescriptionID)
	}
}

func TestRecordRefillRepoError(t *testing.T) {
	recorder := &mockRefillRecorder{err: errors.New("db down")}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Refill: recorder})

	err := svc.RecordRefill(context.Background(), 1, date(2026, 2, 1))
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- List tests ---

func TestListByPatientSuccess(t *testing.T) {
	lister := &mockLister{result: []prescription.Prescription{
		{ID: 1, MedicationName: "Tachipirina"},
		{ID: 2, MedicationName: "Aspirina"},
	}}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Lister: lister})

	got, err := svc.ListByPatient(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

// --- Get tests ---

func TestGetSuccess(t *testing.T) {
	getter := &mockGetter{result: prescription.Prescription{ID: 1, MedicationName: "Tachipirina"}}
	svc := prescription.NewServiceWith(prescription.ServiceDeps{Getter: getter})

	got, err := svc.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MedicationName != "Tachipirina" {
		t.Errorf("MedicationName = %q, want Tachipirina", got.MedicationName)
	}
}

