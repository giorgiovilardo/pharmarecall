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
	"github.com/giorgiovilardo/pharmarecall/internal/prescription"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// PrescriptionLister lists prescriptions for a patient.
type PrescriptionLister interface {
	ListByPatient(ctx context.Context, patientID int64) ([]prescription.Prescription, error)
}

// PrescriptionCreator creates a prescription.
type PrescriptionCreator interface {
	Create(ctx context.Context, p prescription.CreateParams) (prescription.Prescription, error)
}

// PrescriptionGetter fetches a prescription by ID.
type PrescriptionGetter interface {
	Get(ctx context.Context, id int64) (prescription.Prescription, error)
}

// PrescriptionUpdater updates a prescription.
type PrescriptionUpdater interface {
	Update(ctx context.Context, p prescription.UpdateParams) error
}

// PrescriptionRefiller records a refill.
type PrescriptionRefiller interface {
	RecordRefill(ctx context.Context, p prescription.RefillParams) error
}

// HandleNewPrescriptionPage renders the prescription creation form.
func HandleNewPrescriptionPage(patientGetter PatientGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		p, err := patientGetter.Get(r.Context(), patientID)
		if err != nil {
			if errors.Is(err, patient.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting patient", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		if !p.Consensus {
			http.Redirect(w, r, fmt.Sprintf("/patients/%d", patientID), http.StatusSeeOther)
			return
		}

		web.PrescriptionNewPage(p, "").Render(r.Context(), w)
	}
}

// HandleCreatePrescription validates the form and creates a prescription.
func HandleCreatePrescription(creator PrescriptionCreator, patientGetter PatientGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}

		p, err := patientGetter.Get(r.Context(), patientID)
		if err != nil {
			if errors.Is(err, patient.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting patient", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		medicationName := r.FormValue("medication_name")
		unitsPerBox, _ := strconv.Atoi(r.FormValue("units_per_box"))
		dailyConsumption, _ := strconv.ParseFloat(r.FormValue("daily_consumption"), 64)
		boxStartDate, _ := time.Parse("2006-01-02", r.FormValue("box_start_date"))

		renderError := func(errMsg string) {
			web.PrescriptionNewPage(p, errMsg).Render(r.Context(), w)
		}

		if medicationName == "" {
			renderError("Il nome del farmaco è obbligatorio.")
			return
		}
		if unitsPerBox <= 0 {
			renderError("Le unità per confezione devono essere maggiori di zero.")
			return
		}
		if dailyConsumption <= 0 {
			renderError("Il consumo giornaliero deve essere maggiore di zero.")
			return
		}
		if boxStartDate.IsZero() {
			renderError("La data di inizio confezione è obbligatoria.")
			return
		}

		_, err = creator.Create(r.Context(), prescription.CreateParams{
			PatientID:        patientID,
			MedicationName:   medicationName,
			UnitsPerBox:      unitsPerBox,
			DailyConsumption: dailyConsumption,
			BoxStartDate:     boxStartDate,
		})
		if err != nil {
			if errors.Is(err, prescription.ErrNoConsensus) {
				renderError("Il paziente deve dare il consenso prima di aggiungere prescrizioni.")
				return
			}
			slog.Error("creating prescription", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/patients/%d", patientID), http.StatusSeeOther)
	}
}

// HandlePrescriptionEditPage renders the prescription edit form.
func HandlePrescriptionEditPage(prescriptionGetter PrescriptionGetter, patientGetter PatientGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		rxID, err := strconv.ParseInt(r.PathValue("rxid"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		p, err := patientGetter.Get(r.Context(), patientID)
		if err != nil {
			if errors.Is(err, patient.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting patient", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		rx, err := prescriptionGetter.Get(r.Context(), rxID)
		if err != nil {
			if errors.Is(err, prescription.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting prescription", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		web.PrescriptionEditPage(p, rx, "").Render(r.Context(), w)
	}
}

// HandleUpdatePrescription validates the form and updates a prescription.
func HandleUpdatePrescription(updater PrescriptionUpdater, prescriptionGetter PrescriptionGetter, patientGetter PatientGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		rxID, err := strconv.ParseInt(r.PathValue("rxid"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Richiesta non valida.", http.StatusBadRequest)
			return
		}

		p, err := patientGetter.Get(r.Context(), patientID)
		if err != nil {
			if errors.Is(err, patient.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting patient", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		rx, err := prescriptionGetter.Get(r.Context(), rxID)
		if err != nil {
			if errors.Is(err, prescription.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting prescription", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		medicationName := r.FormValue("medication_name")
		unitsPerBox, _ := strconv.Atoi(r.FormValue("units_per_box"))
		dailyConsumption, _ := strconv.ParseFloat(r.FormValue("daily_consumption"), 64)
		boxStartDate, _ := time.Parse("2006-01-02", r.FormValue("box_start_date"))

		renderError := func(errMsg string) {
			web.PrescriptionEditPage(p, rx, errMsg).Render(r.Context(), w)
		}

		if medicationName == "" {
			renderError("Il nome del farmaco è obbligatorio.")
			return
		}
		if unitsPerBox <= 0 {
			renderError("Le unità per confezione devono essere maggiori di zero.")
			return
		}
		if dailyConsumption <= 0 {
			renderError("Il consumo giornaliero deve essere maggiore di zero.")
			return
		}
		if boxStartDate.IsZero() {
			renderError("La data di inizio confezione è obbligatoria.")
			return
		}

		if err := updater.Update(r.Context(), prescription.UpdateParams{
			ID:               rxID,
			MedicationName:   medicationName,
			UnitsPerBox:      unitsPerBox,
			DailyConsumption: dailyConsumption,
			BoxStartDate:     boxStartDate,
		}); err != nil {
			slog.Error("updating prescription", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/patients/%d", patientID), http.StatusSeeOther)
	}
}

// HandleRecordRefill records a refill for a prescription.
func HandleRecordRefill(refiller PrescriptionRefiller) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patientID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		rxID, err := strconv.ParseInt(r.PathValue("rxid"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := refiller.RecordRefill(r.Context(), prescription.RefillParams{
			PrescriptionID: rxID,
			NewStartDate:   time.Now().Truncate(24 * time.Hour),
		}); err != nil {
			slog.Error("recording refill", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/patients/%d", patientID), http.StatusSeeOther)
	}
}
