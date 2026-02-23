package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/patient"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// PatientLister lists patients for a pharmacy.
type PatientLister interface {
	List(ctx context.Context, pharmacyID int64) ([]patient.Summary, error)
}

// PatientCreator creates a patient.
type PatientCreator interface {
	Create(ctx context.Context, p patient.CreateParams) (patient.Patient, error)
}

// PatientGetter fetches a patient by ID.
type PatientGetter interface {
	Get(ctx context.Context, id int64) (patient.Patient, error)
}

// PatientUpdater updates a patient.
type PatientUpdater interface {
	Update(ctx context.Context, p patient.UpdateParams) error
}

// PatientConsensusRecorder records patient consensus.
type PatientConsensusRecorder interface {
	SetConsensus(ctx context.Context, id int64) error
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

// HandleNewPatientPage renders the patient creation form.
func HandleNewPatientPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.PatientNewPage("").Render(r.Context(), w)
	}
}

// patientValidationMessage maps domain validation errors to user-facing messages.
func patientValidationMessage(err error) string {
	switch {
	case errors.Is(err, patient.ErrNameRequired):
		return "Nome e cognome sono obbligatori."
	case errors.Is(err, patient.ErrContactRequired):
		return "È necessario almeno un contatto (telefono o email)."
	case errors.Is(err, patient.ErrDeliveryAddrRequired):
		return "L'indirizzo di consegna è obbligatorio per la spedizione."
	default:
		return ""
	}
}

// HandleCreatePatient parses the form and creates a patient.
func HandleCreatePatient(creator PatientCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.PatientNewPage("Richiesta non valida.").Render(r.Context(), w)
			return
		}

		pharmacyID := web.PharmacyID(r.Context())

		_, err := creator.Create(r.Context(), patient.CreateParams{
			PharmacyID:      pharmacyID,
			FirstName:       r.FormValue("first_name"),
			LastName:        r.FormValue("last_name"),
			Phone:           r.FormValue("phone"),
			Email:           r.FormValue("email"),
			DeliveryAddress: r.FormValue("delivery_address"),
			Fulfillment:     r.FormValue("fulfillment"),
			Notes:           r.FormValue("notes"),
		})
		if err != nil {
			if msg := patientValidationMessage(err); msg != "" {
				web.PatientNewPage(msg).Render(r.Context(), w)
				return
			}
			slog.Error("creating patient", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/patients", http.StatusSeeOther)
	}
}

// HandlePatientDetail renders the patient detail/edit page with prescriptions.
func HandlePatientDetail(getter PatientGetter, rxLister PrescriptionLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		p, err := getter.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, patient.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting patient", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		rxs, err := rxLister.ListByPatient(r.Context(), id)
		if err != nil {
			slog.Error("listing prescriptions", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		web.PatientDetailPage(p, rxs, time.Now(), "").Render(r.Context(), w)
	}
}

// HandleUpdatePatient parses the form and updates a patient.
func HandleUpdatePatient(getter PatientGetter, updater PatientUpdater, rxLister PrescriptionLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}

		renderError := func(errMsg string) {
			p, _ := getter.Get(r.Context(), id)
			rxs, _ := rxLister.ListByPatient(r.Context(), id)
			web.PatientDetailPage(p, rxs, time.Now(), errMsg).Render(r.Context(), w)
		}

		if err := updater.Update(r.Context(), patient.UpdateParams{
			ID:              id,
			FirstName:       r.FormValue("first_name"),
			LastName:        r.FormValue("last_name"),
			Phone:           r.FormValue("phone"),
			Email:           r.FormValue("email"),
			DeliveryAddress: r.FormValue("delivery_address"),
			Fulfillment:     r.FormValue("fulfillment"),
			Notes:           r.FormValue("notes"),
		}); err != nil {
			if msg := patientValidationMessage(err); msg != "" {
				renderError(msg)
				return
			}
			slog.Error("updating patient", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/patients/%d", id), http.StatusSeeOther)
	}
}

// HandleSetConsensus records that a patient has given consensus.
func HandleSetConsensus(recorder PatientConsensusRecorder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := recorder.SetConsensus(r.Context(), id); err != nil {
			slog.Error("setting consensus", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/patients/%d", id), http.StatusSeeOther)
	}
}
