package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/giorgiovilardo/pharmarecall/internal/user"
)

// --- Mocks ---

type mockEmailGetter struct {
	user     user.User
	passHash string
	err      error
}

func (m *mockEmailGetter) GetByEmail(_ context.Context, _ string) (user.User, string, error) {
	return m.user, m.passHash, m.err
}

// --- Tests ---

func TestAuthenticateSuccess(t *testing.T) {
	svc := user.NewServiceWith(user.ServiceDeps{
		EmailGetter: &mockEmailGetter{
			user:     user.User{ID: 1, Email: "admin@example.com", Name: "Admin", Role: "admin"},
			passHash: "hashed-password",
		},
		Verifier: func(hash, password string) error { return nil },
	})

	got, err := svc.Authenticate(context.Background(), "admin@example.com", "secret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if got.Email != "admin@example.com" {
		t.Errorf("Email = %q, want admin@example.com", got.Email)
	}
	if got.Role != "admin" {
		t.Errorf("Role = %q, want admin", got.Role)
	}
}

func TestAuthenticateUserNotFound(t *testing.T) {
	svc := user.NewServiceWith(user.ServiceDeps{
		EmailGetter: &mockEmailGetter{err: user.ErrNotFound},
	})

	_, err := svc.Authenticate(context.Background(), "nobody@example.com", "whatever")
	if !errors.Is(err, user.ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	svc := user.NewServiceWith(user.ServiceDeps{
		EmailGetter: &mockEmailGetter{
			user:     user.User{ID: 1, Email: "user@example.com"},
			passHash: "hashed",
		},
		Verifier: func(_, _ string) error { return errors.New("mismatch") },
	})

	_, err := svc.Authenticate(context.Background(), "user@example.com", "wrong")
	if !errors.Is(err, user.ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

// --- ChangePassword mocks ---

type mockIDGetter struct {
	user     user.User
	passHash string
	err      error
}

func (m *mockIDGetter) GetByID(_ context.Context, _ int64) (user.User, string, error) {
	return m.user, m.passHash, m.err
}

type mockPasswordUpdater struct {
	called  bool
	gotID   int64
	gotHash string
	err     error
}

func (m *mockPasswordUpdater) UpdatePassword(_ context.Context, id int64, hash string) error {
	m.called = true
	m.gotID = id
	m.gotHash = hash
	return m.err
}

// --- ChangePassword tests ---

func TestChangePasswordSuccess(t *testing.T) {
	updater := &mockPasswordUpdater{}
	svc := user.NewServiceWith(user.ServiceDeps{
		IDGetter:        &mockIDGetter{user: user.User{ID: 1}, passHash: "old-hash"},
		PasswordUpdater: updater,
		Verifier:        func(_, _ string) error { return nil },
		Hasher:          func(s string) (string, error) { return "new-hash", nil },
	})

	err := svc.ChangePassword(context.Background(), 1, "old-pass", "new-pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updater.called {
		t.Fatal("UpdatePassword was not called")
	}
	if updater.gotID != 1 {
		t.Errorf("updated ID = %d, want 1", updater.gotID)
	}
	if updater.gotHash != "new-hash" {
		t.Errorf("updated hash = %q, want new-hash", updater.gotHash)
	}
}

func TestChangePasswordWrongCurrent(t *testing.T) {
	svc := user.NewServiceWith(user.ServiceDeps{
		IDGetter: &mockIDGetter{user: user.User{ID: 1}, passHash: "old-hash"},
		Verifier: func(_, _ string) error { return errors.New("mismatch") },
	})

	err := svc.ChangePassword(context.Background(), 1, "wrong", "new-pass")
	if !errors.Is(err, user.ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

// --- SeedAdmin mocks ---

type mockCreator struct {
	called   bool
	gotEmail string
	gotHash  string
	gotName  string
	gotRole  string
	user     user.User
	err      error
}

func (m *mockCreator) Create(_ context.Context, email, passwordHash, name, role string) (user.User, error) {
	m.called = true
	m.gotEmail = email
	m.gotHash = passwordHash
	m.gotName = name
	m.gotRole = role
	return m.user, m.err
}

// --- SeedAdmin tests ---

func TestSeedAdminSuccess(t *testing.T) {
	creator := &mockCreator{user: user.User{ID: 1, Email: "admin@example.com", Name: "Admin", Role: "admin"}}
	svc := user.NewServiceWith(user.ServiceDeps{
		Creator: creator,
		Hasher:  func(s string) (string, error) { return "hashed-" + s, nil },
	})

	got, err := svc.SeedAdmin(context.Background(), "admin@example.com", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if !creator.called {
		t.Fatal("Create was not called")
	}
	if creator.gotEmail != "admin@example.com" {
		t.Errorf("email = %q, want admin@example.com", creator.gotEmail)
	}
	if creator.gotHash != "hashed-secret" {
		t.Errorf("hash = %q, want hashed-secret", creator.gotHash)
	}
	if creator.gotRole != "admin" {
		t.Errorf("role = %q, want admin", creator.gotRole)
	}
	if creator.gotName != "Admin" {
		t.Errorf("name = %q, want Admin", creator.gotName)
	}
}
