package web

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type contextKey string

const (
	ctxKeyUserID     contextKey = "userID"
	ctxKeyRole       contextKey = "role"
	ctxKeyPharmacyID contextKey = "pharmacyID"
)

// LoadUser reads userID and role from the session and attaches them to
// the request context. Does not redirect — use on all routes so the
// layout can conditionally show nav items.
func LoadUser(sessions *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := sessions.GetInt64(r.Context(), "userID")
			if userID == 0 {
				next.ServeHTTP(w, r)
				return
			}

			role := sessions.GetString(r.Context(), "role")
			pharmacyID := sessions.GetInt64(r.Context(), "pharmacyID")

			ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
			ctx = context.WithValue(ctx, ctxKeyRole, role)
			ctx = context.WithValue(ctx, ctxKeyPharmacyID, pharmacyID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth redirects to /login if the user is not loaded in context.
// Must be used after LoadUser.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserID(r.Context()) == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin returns 403 Forbidden if the authenticated user is not an admin.
// Must be used after LoadUser and RequireAuth.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Role(r.Context()) != "admin" {
			http.Error(w, "Accesso negato.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePharmacyStaff returns 403 Forbidden if the authenticated user has no
// pharmacy association (i.e., is not an owner or personnel).
// Must be used after LoadUser and RequireAuth.
func RequirePharmacyStaff(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if PharmacyID(r.Context()) == 0 {
			http.Error(w, "Accesso negato.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireOwner returns 403 Forbidden if the authenticated user is not an owner.
// Must be used after LoadUser and RequireAuth.
func RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if Role(r.Context()) != "owner" {
			http.Error(w, "Accesso negato.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// UserID returns the authenticated user's ID from the request context.
func UserID(ctx context.Context) int64 {
	id, _ := ctx.Value(ctxKeyUserID).(int64)
	return id
}

// Role returns the authenticated user's role from the request context.
func Role(ctx context.Context) string {
	role, _ := ctx.Value(ctxKeyRole).(string)
	return role
}

// PharmacyID returns the authenticated user's pharmacy ID from the request context.
func PharmacyID(ctx context.Context) int64 {
	id, _ := ctx.Value(ctxKeyPharmacyID).(int64)
	return id
}
