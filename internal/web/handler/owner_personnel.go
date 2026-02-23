package handler

import (
	"log/slog"
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// HandleOwnerPersonnelList renders the personnel list for the pharmacy owner.
func HandleOwnerPersonnelList(lister PersonnelLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID := web.PharmacyID(r.Context())

		members, err := lister.ListPersonnel(r.Context(), pharmacyID)
		if err != nil {
			slog.Error("listing personnel", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		web.OwnerPersonnelPage(members).Render(r.Context(), w)
	}
}
