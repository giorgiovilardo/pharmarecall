package user

import (
	"context"
	"errors"
	"fmt"
)

// ServiceDeps holds individual port interfaces — used by tests to inject only what's needed.
type ServiceDeps struct {
	EmailGetter     UserByEmailGetter
	IDGetter        UserByIDGetter
	PasswordUpdater PasswordUpdater
	Creator         UserCreator
	Hasher          func(string) (string, error)
	Verifier        func(hash, password string) error
}

// Service contains user domain business logic.
type Service struct {
	deps ServiceDeps
}

// NewService is the production constructor — takes a Repository (satisfies all ports).
func NewService(repo Repository, hasher func(string) (string, error), verifier func(hash, password string) error) *Service {
	return &Service{deps: ServiceDeps{
		EmailGetter:     repo,
		IDGetter:        repo,
		PasswordUpdater: repo,
		Creator:         repo,
		Hasher:          hasher,
		Verifier:        verifier,
	}}
}

// NewServiceWith is the test constructor — inject only what you need, rest stays nil.
func NewServiceWith(d ServiceDeps) *Service {
	return &Service{deps: d}
}

// Authenticate verifies credentials and returns the user.
func (s *Service) Authenticate(ctx context.Context, email, password string) (User, error) {
	u, hash, err := s.deps.EmailGetter.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, fmt.Errorf("looking up user: %w", err)
	}

	if err := s.deps.Verifier(hash, password); err != nil {
		return User{}, ErrInvalidCredentials
	}

	return u, nil
}

// ChangePassword verifies the current password and updates to the new one.
func (s *Service) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	_, hash, err := s.deps.IDGetter.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("looking up user: %w", err)
	}

	if err := s.deps.Verifier(hash, currentPassword); err != nil {
		return ErrInvalidCredentials
	}

	newHash, err := s.deps.Hasher(newPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	if err := s.deps.PasswordUpdater.UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}

// SeedAdmin creates an admin user with the given email and password.
func (s *Service) SeedAdmin(ctx context.Context, email, password string) (User, error) {
	hash, err := s.deps.Hasher(password)
	if err != nil {
		return User{}, fmt.Errorf("hashing admin password: %w", err)
	}

	u, err := s.deps.Creator.Create(ctx, email, hash, "Admin", "admin")
	if err != nil {
		return User{}, fmt.Errorf("creating admin user: %w", err)
	}

	return u, nil
}
