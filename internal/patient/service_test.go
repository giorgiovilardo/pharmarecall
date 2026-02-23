package patient_test

import (
	"context"
	"errors"
	"testing"

	"github.com/giorgiovilardo/pharmarecall/internal/patient"
)

// --- Mocks ---

type mockPatientCreator struct {
	called bool
	result patient.Patient
	err    error
}

func (m *mockPatientCreator) Create(_ context.Context, _ patient.CreateParams) (patient.Patient, error) {
	m.called = true
	return m.result, m.err
}

type mockPatientUpdater struct {
	called bool
	err    error
}

func (m *mockPatientUpdater) Update(_ context.Context, _ patient.UpdateParams) error {
	m.called = true
	return m.err
}

// --- Create tests ---

func TestCreateSuccess(t *testing.T) {
	creator := &mockPatientCreator{result: patient.Patient{ID: 1, FirstName: "Mario", LastName: "Rossi"}}
	svc := patient.NewServiceWith(patient.ServiceDeps{Creator: creator})

	got, err := svc.Create(context.Background(), patient.CreateParams{
		FirstName: "Mario",
		LastName:  "Rossi",
		Phone:     "333-1234567",
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

func TestCreateDefaultsFulfillmentToPickup(t *testing.T) {
	creator := &mockPatientCreator{result: patient.Patient{ID: 1}}
	svc := patient.NewServiceWith(patient.ServiceDeps{Creator: creator})

	_, err := svc.Create(context.Background(), patient.CreateParams{
		FirstName: "Mario",
		LastName:  "Rossi",
		Phone:     "333-1234567",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !creator.called {
		t.Fatal("Create was not called")
	}
}

func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name   string
		params patient.CreateParams
		errStr string
	}{
		{
			name:   "missing first name",
			params: patient.CreateParams{LastName: "Rossi", Phone: "333"},
			errStr: "nome",
		},
		{
			name:   "missing last name",
			params: patient.CreateParams{FirstName: "Mario", Phone: "333"},
			errStr: "cognome",
		},
		{
			name:   "missing contact",
			params: patient.CreateParams{FirstName: "Mario", LastName: "Rossi"},
			errStr: "contatto",
		},
		{
			name:   "shipping without address",
			params: patient.CreateParams{FirstName: "Mario", LastName: "Rossi", Phone: "333", Fulfillment: "shipping"},
			errStr: "indirizzo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := patient.NewServiceWith(patient.ServiceDeps{})
			_, err := svc.Create(context.Background(), tt.params)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !containsSubstring(err.Error(), tt.errStr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.errStr)
			}
		})
	}
}

// --- Update tests ---

func TestUpdateSuccess(t *testing.T) {
	updater := &mockPatientUpdater{}
	svc := patient.NewServiceWith(patient.ServiceDeps{Updater: updater})

	err := svc.Update(context.Background(), patient.UpdateParams{
		ID: 1, FirstName: "Mario", LastName: "Rossi", Phone: "333",
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
		params patient.UpdateParams
		errStr string
	}{
		{
			name:   "missing name",
			params: patient.UpdateParams{ID: 1, LastName: "Rossi", Phone: "333"},
			errStr: "nome",
		},
		{
			name:   "missing contact",
			params: patient.UpdateParams{ID: 1, FirstName: "Mario", LastName: "Rossi"},
			errStr: "contatto",
		},
		{
			name:   "shipping without address",
			params: patient.UpdateParams{ID: 1, FirstName: "Mario", LastName: "Rossi", Phone: "333", Fulfillment: "shipping"},
			errStr: "indirizzo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockPatientUpdater{}
			svc := patient.NewServiceWith(patient.ServiceDeps{Updater: updater})
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

// --- SetConsensus tests ---

type mockConsensusRecorder struct {
	called bool
	err    error
}

func (m *mockConsensusRecorder) SetConsensus(_ context.Context, _ int64) error {
	m.called = true
	return m.err
}

func TestSetConsensusSuccess(t *testing.T) {
	recorder := &mockConsensusRecorder{}
	svc := patient.NewServiceWith(patient.ServiceDeps{Consensus: recorder})

	if err := svc.SetConsensus(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !recorder.called {
		t.Fatal("SetConsensus was not called")
	}
}

func TestSetConsensusRepoError(t *testing.T) {
	recorder := &mockConsensusRecorder{err: errors.New("db down")}
	svc := patient.NewServiceWith(patient.ServiceDeps{Consensus: recorder})

	err := svc.SetConsensus(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
