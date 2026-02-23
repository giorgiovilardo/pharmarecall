package user

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("user not found")
)

// User is the domain representation of a user.
type User struct {
	ID         int64
	Email      string
	Name       string
	Role       string
	PharmacyID int64 // 0 means no pharmacy (admin users)
}
