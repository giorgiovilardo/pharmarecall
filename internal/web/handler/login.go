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

// Authenticator verifies credentials and returns a user.
type Authenticator interface {
	Authenticate(ctx context.Context, email, password string) (user.User, error)
}

// HandleLoginPage renders the login form.
func HandleLoginPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.LoginPage("").Render(r.Context(), w)
	}
}

// HandleLoginPost validates credentials, creates a session, and redirects.
func HandleLoginPost(sessions *scs.SessionManager, auth Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.LoginPage("Richiesta non valida.").Render(r.Context(), w)
			return
		}

		u, err := auth.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
		if err != nil {
			if errors.Is(err, user.ErrInvalidCredentials) {
				web.LoginPage("Credenziali non valide.").Render(r.Context(), w)
				return
			}
			slog.Error("authenticating", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		if err := sessions.RenewToken(r.Context()); err != nil {
			slog.Error("renewing session token", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		sessions.Put(r.Context(), "userID", u.ID)
		sessions.Put(r.Context(), "role", u.Role)

		dest := "/dashboard"
		if u.Role == "admin" {
			dest = "/admin"
		}

		http.Redirect(w, r, dest, http.StatusSeeOther)
	}
}
