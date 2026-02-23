package web

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/giorgiovilardo/pharmarecall/internal/auth"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreatePersonnelFunc creates a personnel user scoped to a pharmacy, transactionally.
type CreatePersonnelFunc func(ctx context.Context, arg db.CreateUserParams) (db.User, error)

// HandleAddPersonnelPage renders the add-personnel form for a pharmacy.
func HandleAddPersonnelPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		AddPersonnelPage(id, "").Render(r.Context(), w)
	}
}

// HandleCreatePersonnel creates a new personnel user scoped to a pharmacy.
func HandleCreatePersonnel(createFn CreatePersonnelFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			AddPersonnelPage(pharmacyID, "Richiesta non valida.").Render(r.Context(), w)
			return
		}

		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if name == "" || email == "" || password == "" {
			AddPersonnelPage(pharmacyID, "Tutti i campi sono obbligatori.").Render(r.Context(), w)
			return
		}

		hash, err := auth.HashPassword(password)
		if err != nil {
			slog.Error("hashing personnel password", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		role := "personnel"
		if r.FormValue("owner") == "true" {
			role = "owner"
		}

		_, err = createFn(r.Context(), db.CreateUserParams{
			Email:        email,
			PasswordHash: hash,
			Name:         name,
			Role:         role,
			PharmacyID:   pgtype.Int8{Int64: pharmacyID, Valid: true},
		})
		if err != nil {
			if isDuplicateEmail(err) {
				AddPersonnelPage(pharmacyID, "L'email è già in uso.").Render(r.Context(), w)
				return
			}
			slog.Error("creating personnel user", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin/pharmacies/"+strconv.FormatInt(pharmacyID, 10), http.StatusSeeOther)
	}
}
