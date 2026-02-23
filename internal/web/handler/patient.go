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

// PatientCreator creates a patient.
type PatientCreator interface {
	Create(ctx context.Context, p patient.CreateParams) (patient.Patient, error)
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
