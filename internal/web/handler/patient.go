package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/internal/patient"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// PatientLister lists patients for a pharmacy.
type PatientLister interface {
	List(ctx context.Context, pharmacyID int64) ([]patient.Summary, error)
}

// HandlePatientList renders the patient list page.
func HandlePatientList(lister PatientLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID := web.PharmacyID(r.Context())

		patients, err := lister.List(r.Context(), pharmacyID)
		if err != nil {
			slog.Error("listing patients", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		web.PatientListPage(patients).Render(r.Context(), w)
	}
}
