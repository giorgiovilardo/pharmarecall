package web

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

type contextKey string

const (
	ctxKeyUserID contextKey = "userID"
	ctxKeyRole   contextKey = "role"
)

// RequireAuth redirects to /login if the session has no userID.
// Otherwise it attaches userID and role to the request context.
func RequireAuth(sessions *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := sessions.GetInt64(r.Context(), "userID")
			if userID == 0 {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			role := sessions.GetString(r.Context(), "role")

			ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
			ctx = context.WithValue(ctx, ctxKeyRole, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
