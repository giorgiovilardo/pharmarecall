package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/user"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// PasswordChanger changes a user's password.
type PasswordChanger interface {
	ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error
}

// HandleChangePasswordPage renders the change password form.
func HandleChangePasswordPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.ChangePasswordPage("", "").Render(r.Context(), w)
	}
}

// HandleChangePasswordPost verifies the current password and updates to the new one.
func HandleChangePasswordPost(sessions *scs.SessionManager, changer PasswordChanger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.ChangePasswordPage("Richiesta non valida.", "").Render(r.Context(), w)
			return
		}

		userID := sessions.GetInt64(r.Context(), "userID")

		err := changer.ChangePassword(r.Context(), userID, r.FormValue("current_password"), r.FormValue("new_password"))
		if err != nil {
			if errors.Is(err, user.ErrInvalidCredentials) {
				web.ChangePasswordPage("Password attuale non corretta.", "").Render(r.Context(), w)
				return
			}
			slog.Error("changing password", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		web.ChangePasswordPage("", "Password aggiornata.").Render(r.Context(), w)
	}
}
