package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
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

// HandleOwnerAddPersonnelPage renders the add-personnel form for a pharmacy owner.
func HandleOwnerAddPersonnelPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		web.OwnerAddPersonnelPage("").Render(r.Context(), w)
	}
}

// HandleOwnerCreatePersonnel creates a new personnel user scoped to the owner's pharmacy.
func HandleOwnerCreatePersonnel(creator PersonnelCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			web.OwnerAddPersonnelPage("Richiesta non valida.").Render(r.Context(), w)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if name == "" || email == "" || password == "" {
			web.OwnerAddPersonnelPage("Tutti i campi sono obbligatori.").Render(r.Context(), w)
			return
		}

		pharmacyID := web.PharmacyID(r.Context())

		_, err := creator.CreatePersonnel(r.Context(), pharmacy.CreatePersonnelParams{
			PharmacyID: pharmacyID,
			Name:       name,
			Email:      email,
			Password:   password,
			Role:       "personnel",
		})
		if err != nil {
			if errors.Is(err, pharmacy.ErrDuplicateEmail) {
				web.OwnerAddPersonnelPage("L'email è già in uso.").Render(r.Context(), w)
				return
			}
			slog.Error("creating personnel user", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/personnel", http.StatusSeeOther)
	}
}
