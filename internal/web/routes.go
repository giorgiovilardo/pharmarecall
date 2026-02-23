package web

import (
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/static"
)

// NewRouter builds the ServeMux with all routes. Handlers are constructed
// by the caller (main or tests) and passed in ready to use.
// Middleware (sessions, CORS) is applied by the caller.
func NewRouter(loginPage, loginPost, logoutPost, changePasswordPage, changePasswordPost http.HandlerFunc) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.Files)))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		HealthPage().Render(r.Context(), w)
	})
	mux.HandleFunc("GET /login", loginPage)
	mux.HandleFunc("POST /login", loginPost)
	mux.HandleFunc("POST /logout", logoutPost)
	mux.HandleFunc("GET /change-password", changePasswordPage)
	mux.HandleFunc("POST /change-password", changePasswordPost)
	return mux
}
