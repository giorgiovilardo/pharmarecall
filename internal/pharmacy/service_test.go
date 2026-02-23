package pharmacy_test

import (
	"context"
	"testing"

	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
)

// --- Mocks ---

type mockPharmacyCreator struct {
	called  bool
	gotHash string
	result  pharmacy.Pharmacy
	err     error
}

func (m *mockPharmacyCreator) CreateWithOwner(_ context.Context, _ pharmacy.CreateParams, hash string) (pharmacy.Pharmacy, error) {
	m.called = true
	m.gotHash = hash
	return m.result, m.err
}

// --- CreateWithOwner tests ---

func TestCreateWithOwnerSuccess(t *testing.T) {
	creator := &mockPharmacyCreator{result: pharmacy.Pharmacy{ID: 1, Name: "Farmacia Rossi"}}
	svc := pharmacy.NewServiceWith(pharmacy.ServiceDeps{
		Creator: creator,
		Hasher:  func(s string) (string, error) { return "hashed-" + s, nil },
	})

	got, err := svc.CreateWithOwner(context.Background(), pharmacy.CreateParams{
		Name:          "Farmacia Rossi",
		Address:       "Via Roma 1",
		OwnerName:     "Mario",
		OwnerEmail:    "mario@example.com",
		OwnerPassword: "secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if !creator.called {
		t.Fatal("CreateWithOwner was not called")
	}
	if creator.gotHash != "hashed-secret" {
		t.Errorf("hash = %q, want hashed-secret", creator.gotHash)
	}
}

// --- CreatePersonnel mocks ---

type mockPersonnelCreator struct {
	called  bool
	gotHash string
	result  pharmacy.PersonnelMember
	err     error
}

func (m *mockPersonnelCreator) CreatePersonnel(_ context.Context, _ pharmacy.CreatePersonnelParams, hash string) (pharmacy.PersonnelMember, error) {
	m.called = true
	m.gotHash = hash
	return m.result, m.err
}

// --- CreatePersonnel tests ---

func TestCreatePersonnelSuccess(t *testing.T) {
	creator := &mockPersonnelCreator{result: pharmacy.PersonnelMember{ID: 5, Name: "Anna", Email: "anna@example.com", Role: "personnel"}}
	svc := pharmacy.NewServiceWith(pharmacy.ServiceDeps{
		PersCreator: creator,
		Hasher:      func(s string) (string, error) { return "hashed-" + s, nil },
	})

	got, err := svc.CreatePersonnel(context.Background(), pharmacy.CreatePersonnelParams{
		PharmacyID: 1,
		Name:       "Anna",
		Email:      "anna@example.com",
		Password:   "temppass",
		Role:       "personnel",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 5 {
		t.Errorf("ID = %d, want 5", got.ID)
	}
	if !creator.called {
		t.Fatal("CreatePersonnel was not called")
	}
	if creator.gotHash != "hashed-temppass" {
		t.Errorf("hash = %q, want hashed-temppass", creator.gotHash)
	}
}
