package web

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/auth"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
)

// PasswordChanger reads a user and updates their password.
type PasswordChanger interface {
	GetUserByID(ctx context.Context, id int64) (db.User, error)
	UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error
}

// HandleChangePasswordPage renders the change password form.
func HandleChangePasswordPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ChangePasswordPage("", "").Render(r.Context(), w)
	}
}

// HandleChangePasswordPost verifies the current password, hashes the new one, and updates the user.
func HandleChangePasswordPost(sessions *scs.SessionManager, users PasswordChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			ChangePasswordPage("Richiesta non valida.", "").Render(r.Context(), w)
			return
		}

		userID := sessions.GetInt64(r.Context(), "userID")
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")

		user, err := users.GetUserByID(r.Context(), userID)
		if err != nil {
			slog.Error("looking up user for password change", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		if err := auth.VerifyPassword(user.PasswordHash, currentPassword); err != nil {
			ChangePasswordPage("Password attuale non corretta.", "").Render(r.Context(), w)
			return
		}

		hash, err := auth.HashPassword(newPassword)
		if err != nil {
			slog.Error("hashing new password", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		if err := users.UpdateUserPassword(r.Context(), db.UpdateUserPasswordParams{
			ID:           userID,
			PasswordHash: hash,
		}); err != nil {
			slog.Error("updating password", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		ChangePasswordPage("", "Password aggiornata.").Render(r.Context(), w)
	}
}
