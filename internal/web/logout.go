package web

import (
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

// HandleLogout destroys the session and redirects to login.
func HandleLogout(sessions *scs.SessionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := sessions.Destroy(r.Context()); err != nil {
			slog.Error("destroying session", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
