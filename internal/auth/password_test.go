package auth_test

import (
	"testing"

	"github.com/giorgiovilardo/pharmarecall/internal/auth"
)

func TestHashAndVerifyPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}

	if err := auth.VerifyPassword(hash, "correct-password"); err != nil {
		t.Errorf("verifying correct password: %v", err)
	}

	if err := auth.VerifyPassword(hash, "wrong-password"); err == nil {
		t.Error("expected error for wrong password, got nil")
	}
}

func TestHashPasswordProducesDifferentHashes(t *testing.T) {
	h1, err := auth.HashPassword("same-password")
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}

	h2, err := auth.HashPassword("same-password")
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}

	if h1 == h2 {
		t.Error("expected different hashes for same password (bcrypt uses random salt)")
	}
}
