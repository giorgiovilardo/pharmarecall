package web

import (
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/static"
)

// NewRouter builds the ServeMux with all routes. Handlers are constructed
// by the caller (main or tests) and passed in ready to use.
// Middleware (sessions, CORS) is applied by the caller.
// AdminHandlers groups all admin-only handler funcs.
type AdminHandlers struct {
	Dashboard       http.HandlerFunc
	NewPharmacy     http.HandlerFunc
	CreatePharmacy  http.HandlerFunc
	PharmacyDetail  http.HandlerFunc
	UpdatePharmacy  http.HandlerFunc
	AddPersonnel    http.HandlerFunc
	CreatePersonnel http.HandlerFunc
}

func NewRouter(loginPage, loginPost, logoutPost, changePasswordPage, changePasswordPost http.HandlerFunc, admin AdminHandlers) *http.ServeMux {
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

	// Admin routes — RequireAdmin middleware applied per-handler
	mux.Handle("GET /admin", RequireAdmin(http.HandlerFunc(admin.Dashboard)))
	mux.Handle("GET /admin/pharmacies/new", RequireAdmin(http.HandlerFunc(admin.NewPharmacy)))
	mux.Handle("POST /admin/pharmacies", RequireAdmin(http.HandlerFunc(admin.CreatePharmacy)))
	mux.Handle("GET /admin/pharmacies/{id}", RequireAdmin(http.HandlerFunc(admin.PharmacyDetail)))
	mux.Handle("POST /admin/pharmacies/{id}", RequireAdmin(http.HandlerFunc(admin.UpdatePharmacy)))
	mux.Handle("GET /admin/pharmacies/{id}/personnel/new", RequireAdmin(http.HandlerFunc(admin.AddPersonnel)))
	mux.Handle("POST /admin/pharmacies/{id}/personnel", RequireAdmin(http.HandlerFunc(admin.CreatePersonnel)))

	return mux
}
