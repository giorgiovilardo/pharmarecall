package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// PharmacyLister lists all pharmacies with personnel counts.
type PharmacyLister interface {
	List(ctx context.Context) ([]pharmacy.Summary, error)
}

// HandleAdminDashboard renders the admin dashboard with the pharmacy list.
func HandleAdminDashboard(lister PharmacyLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := lister.List(r.Context())
		if err != nil {
			slog.Error("listing pharmacies", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}
		web.AdminDashboardPage(rows).Render(r.Context(), w)
	}
}
