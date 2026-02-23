package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// PharmacyGetter fetches a pharmacy by ID.
type PharmacyGetter interface {
	Get(ctx context.Context, id int64) (pharmacy.Pharmacy, error)
}

// PersonnelLister lists personnel for a pharmacy.
type PersonnelLister interface {
	ListPersonnel(ctx context.Context, pharmacyID int64) ([]pharmacy.PersonnelMember, error)
}

// PharmacyUpdater updates a pharmacy.
type PharmacyUpdater interface {
	Update(ctx context.Context, p pharmacy.UpdateParams) error
}

// PharmacyCreatorWithOwner creates a pharmacy with its owner.
type PharmacyCreatorWithOwner interface {
	CreateWithOwner(ctx context.Context, p pharmacy.CreateParams) (pharmacy.Pharmacy, error)
}

// HandlePharmacyDetail renders the pharmacy detail/edit page for admin.
func HandlePharmacyDetail(getter PharmacyGetter, personnel PersonnelLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		p, err := getter.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, pharmacy.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			slog.Error("getting pharmacy", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		members, err := personnel.ListPersonnel(r.Context(), id)
		if err != nil {
			slog.Error("listing pharmacy personnel", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		web.PharmacyDetailPage(p, members, "").Render(r.Context(), w)
	}
}

// HandleUpdatePharmacy handles the form POST to update pharmacy details.
func HandleUpdatePharmacy(getter PharmacyGetter, personnel PersonnelLister, updater PharmacyUpdater) http.HandlerFunc {
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
			p, _ := getter.Get(r.Context(), id)
			members, _ := personnel.ListPersonnel(r.Context(), id)
			web.PharmacyDetailPage(p, members, "Nome e indirizzo sono obbligatori.").Render(r.Context(), w)
			return
		}

		if err := updater.Update(r.Context(), pharmacy.UpdateParams{
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

// HandleNewPharmacyPage renders the pharmacy creation form.
func HandleNewPharmacyPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.NewPharmacyPage("").Render(r.Context(), w)
	}
}

// HandleCreatePharmacy validates the form and creates a pharmacy with owner.
func HandleCreatePharmacy(creator PharmacyCreatorWithOwner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.NewPharmacyPage("Richiesta non valida.").Render(r.Context(), w)
			return
		}

		pharmacyName := r.FormValue("pharmacy_name")
		address := r.FormValue("address")
		phone := r.FormValue("phone")
		email := r.FormValue("email")
		ownerName := r.FormValue("owner_name")
		ownerEmail := r.FormValue("owner_email")
		ownerPassword := r.FormValue("owner_password")

		if pharmacyName == "" || address == "" || ownerName == "" || ownerEmail == "" || ownerPassword == "" {
			web.NewPharmacyPage("Tutti i campi obbligatori devono essere compilati.").Render(r.Context(), w)
			return
		}

		_, err := creator.CreateWithOwner(r.Context(), pharmacy.CreateParams{
			Name:          pharmacyName,
			Address:       address,
			Phone:         phone,
			Email:         email,
			OwnerName:     ownerName,
			OwnerEmail:    ownerEmail,
			OwnerPassword: ownerPassword,
		})
		if err != nil {
			if errors.Is(err, pharmacy.ErrDuplicateEmail) {
				web.NewPharmacyPage("L'email del titolare è già in uso.").Render(r.Context(), w)
				return
			}
			slog.Error("creating pharmacy with owner", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}
