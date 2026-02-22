package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/auth"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5"
)

// UserByEmailGetter looks up a user by email address.
type UserByEmailGetter interface {
	GetUserByEmail(ctx context.Context, email string) (db.User, error)
}

// HandleLoginPage renders the login form.
func HandleLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		LoginPage("").Render(r.Context(), w)
	}
}

// HandleLoginPost validates credentials, creates a session, and redirects.
func HandleLoginPost(sessions *scs.SessionManager, users UserByEmailGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			LoginPage("Richiesta non valida.").Render(r.Context(), w)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := users.GetUserByEmail(r.Context(), email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				LoginPage("Credenziali non valide.").Render(r.Context(), w)
				return
			}
			slog.Error("looking up user", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
			LoginPage("Credenziali non valide.").Render(r.Context(), w)
			return
		}

		if err := sessions.RenewToken(r.Context()); err != nil {
			slog.Error("renewing session token", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		sessions.Put(r.Context(), "userID", user.ID)
		sessions.Put(r.Context(), "role", user.Role)

		dest := "/dashboard"
		if user.Role == "admin" {
			dest = "/admin"
		}

		http.Redirect(w, r, dest, http.StatusSeeOther)
	}
}
