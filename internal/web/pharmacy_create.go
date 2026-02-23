package web

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/internal/auth"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreatePharmacyWithOwnerFunc creates a pharmacy and its owner user transactionally.
type CreatePharmacyWithOwnerFunc func(ctx context.Context, p db.CreatePharmacyParams, owner db.CreateUserParams) (db.Pharmacy, error)

// HandleNewPharmacyPage renders the pharmacy creation form.
func HandleNewPharmacyPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		NewPharmacyPage("").Render(r.Context(), w)
	}
}

// HandleCreatePharmacy validates the form and delegates to createFn for transactional persistence.
func HandleCreatePharmacy(createFn CreatePharmacyWithOwnerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			NewPharmacyPage("Richiesta non valida.").Render(r.Context(), w)
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
			NewPharmacyPage("Tutti i campi obbligatori devono essere compilati.").Render(r.Context(), w)
			return
		}

		hash, err := auth.HashPassword(ownerPassword)
		if err != nil {
			slog.Error("hashing owner password", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		_, err = createFn(r.Context(), db.CreatePharmacyParams{
			Name:    pharmacyName,
			Address: address,
			Phone:   phone,
			Email:   email,
		}, db.CreateUserParams{
			Email:        ownerEmail,
			PasswordHash: hash,
			Name:         ownerName,
			Role:         "owner",
		})
		if err != nil {
			if isDuplicateEmail(err) {
				NewPharmacyPage("L'email del titolare è già in uso.").Render(r.Context(), w)
				return
			}
			slog.Error("creating pharmacy with owner", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}

// isDuplicateEmail checks if a pgx error is a unique constraint violation.
func isDuplicateEmail(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
