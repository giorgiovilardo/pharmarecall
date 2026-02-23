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

// HandleCreatePatient validates the form and creates a patient.
func HandleCreatePatient(creator PatientCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.PatientNewPage("Richiesta non valida.").Render(r.Context(), w)
			return
		}

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		deliveryAddress := r.FormValue("delivery_address")
		fulfillment := r.FormValue("fulfillment")
		notes := r.FormValue("notes")

		if firstName == "" || lastName == "" {
			web.PatientNewPage("Nome e cognome sono obbligatori.").Render(r.Context(), w)
			return
		}
		if phone == "" && email == "" {
			web.PatientNewPage("È necessario almeno un contatto (telefono o email).").Render(r.Context(), w)
			return
		}
		if fulfillment == "" {
			fulfillment = patient.FulfillmentPickup
		}
		if fulfillment == patient.FulfillmentShipping && deliveryAddress == "" {
			web.PatientNewPage("L'indirizzo di consegna è obbligatorio per la spedizione.").Render(r.Context(), w)
			return
		}

		pharmacyID := web.PharmacyID(r.Context())

		_, err := creator.Create(r.Context(), patient.CreateParams{
			PharmacyID:      pharmacyID,
			FirstName:       firstName,
			LastName:        lastName,
			Phone:           phone,
			Email:           email,
			DeliveryAddress: deliveryAddress,
			Fulfillment:     fulfillment,
			Notes:           notes,
		})
		if err != nil {
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

// HandleUpdatePatient validates the form and updates a patient.
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

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		deliveryAddress := r.FormValue("delivery_address")
		fulfillment := r.FormValue("fulfillment")
		notes := r.FormValue("notes")

		renderError := func(errMsg string) {
			p, _ := getter.Get(r.Context(), id)
			rxs, _ := rxLister.ListByPatient(r.Context(), id)
			web.PatientDetailPage(p, rxs, time.Now(), errMsg).Render(r.Context(), w)
		}

		if firstName == "" || lastName == "" {
			renderError("Nome e cognome sono obbligatori.")
			return
		}
		if phone == "" && email == "" {
			renderError("È necessario almeno un contatto (telefono o email).")
			return
		}
		if fulfillment == patient.FulfillmentShipping && deliveryAddress == "" {
			renderError("L'indirizzo di consegna è obbligatorio per la spedizione.")
			return
		}

		if err := updater.Update(r.Context(), patient.UpdateParams{
			ID:              id,
			FirstName:       firstName,
			LastName:        lastName,
			Phone:           phone,
			Email:           email,
			DeliveryAddress: deliveryAddress,
			Fulfillment:     fulfillment,
			Notes:           notes,
		}); err != nil {
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
