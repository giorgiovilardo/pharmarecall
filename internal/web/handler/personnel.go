package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// PersonnelCreator creates a personnel member for a pharmacy.
type PersonnelCreator interface {
	CreatePersonnel(ctx context.Context, p pharmacy.CreatePersonnelParams) (pharmacy.PersonnelMember, error)
}

// HandleAddPersonnelPage renders the add-personnel form for a pharmacy.
func HandleAddPersonnelPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		web.AddPersonnelPage(id, "").Render(r.Context(), w)
	}
}

// HandleCreatePersonnel creates a new personnel user scoped to a pharmacy.
func HandleCreatePersonnel(creator PersonnelCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			web.AddPersonnelPage(pharmacyID, "Richiesta non valida.").Render(r.Context(), w)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if name == "" || email == "" || password == "" {
			web.AddPersonnelPage(pharmacyID, "Tutti i campi sono obbligatori.").Render(r.Context(), w)
			return
		}

		role := "personnel"
		if r.FormValue("owner") == "true" {
			role = "owner"
		}

		_, err = creator.CreatePersonnel(r.Context(), pharmacy.CreatePersonnelParams{
			PharmacyID: pharmacyID,
			Name:       name,
			Email:      email,
			Password:   password,
			Role:       role,
		})
		if err != nil {
			if errors.Is(err, pharmacy.ErrDuplicateEmail) {
				web.AddPersonnelPage(pharmacyID, "L'email è già in uso.").Render(r.Context(), w)
				return
			}
			slog.Error("creating personnel user", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/pharmacies/"+strconv.FormatInt(pharmacyID, 10), http.StatusSeeOther)
	}
}
