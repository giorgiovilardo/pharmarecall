package web

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
)

// PharmacyLister lists all pharmacies with personnel counts.
type PharmacyLister interface {
	ListPharmacies(ctx context.Context) ([]db.ListPharmaciesRow, error)
}

// HandleAdminDashboard renders the admin dashboard with the pharmacy list.
func HandleAdminDashboard(pharmacies PharmacyLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pharmacies.ListPharmacies(r.Context())
		if err != nil {
			slog.Error("listing pharmacies", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}
		AdminDashboardPage(rows).Render(r.Context(), w)
	}
}
