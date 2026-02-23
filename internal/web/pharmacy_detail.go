package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5"
)

// PharmacyDetailReader reads a pharmacy and its personnel.
type PharmacyDetailReader interface {
	GetPharmacyByID(ctx context.Context, id int64) (db.Pharmacy, error)
	ListUsersByPharmacy(ctx context.Context, pharmacyID int64) ([]db.User, error)
}

// UpdatePharmacyFunc updates a pharmacy in a transaction.
type UpdatePharmacyFunc func(ctx context.Context, arg db.UpdatePharmacyParams) error

// HandlePharmacyDetail renders the pharmacy detail/edit page for admin.
func HandlePharmacyDetail(store PharmacyDetailReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		pharmacy, err := store.GetPharmacyByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting pharmacy", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		personnel, err := store.ListUsersByPharmacy(r.Context(), id)
		if err != nil {
			slog.Error("listing pharmacy personnel", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		PharmacyDetailPage(pharmacy, personnel, "").Render(r.Context(), w)
	}
}

// HandleUpdatePharmacy handles the form POST to update pharmacy details.
func HandleUpdatePharmacy(store PharmacyDetailReader, updateFn UpdatePharmacyFunc) http.HandlerFunc {
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

		name := r.FormValue("name")
		address := r.FormValue("address")
		phone := r.FormValue("phone")
		email := r.FormValue("email")

		if name == "" || address == "" {
			pharmacy, _ := store.GetPharmacyByID(r.Context(), id)
			personnel, _ := store.ListUsersByPharmacy(r.Context(), id)
			PharmacyDetailPage(pharmacy, personnel, "Nome e indirizzo sono obbligatori.").Render(r.Context(), w)
			return
		}

		if err := updateFn(r.Context(), db.UpdatePharmacyParams{
			ID:      id,
			Name:    name,
			Address: address,
			Phone:   phone,
			Email:   email,
		}); err != nil {
			slog.Error("updating pharmacy", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/pharmacies/%d", id), http.StatusSeeOther)
	}
}
