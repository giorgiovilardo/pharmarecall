package web

import (
	"net/http"

	"github.com/giorgiovilardo/pharmarecall/static"
)

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

// OwnerHandlers groups all owner-only handler funcs.
type OwnerHandlers struct {
	PersonnelList http.HandlerFunc
}

// Handlers groups all handler funcs for routing.
type Handlers struct {
	LoginPage      http.HandlerFunc
	LoginPost      http.HandlerFunc
	Logout         http.HandlerFunc
	ChangePassPage http.HandlerFunc
	ChangePassPost http.HandlerFunc
	Admin          AdminHandlers
	Owner          OwnerHandlers
}

// NewRouter builds the ServeMux with all routes. Handlers are constructed
// by the caller (main or tests) and passed in ready to use.
// Middleware (sessions, CORS) is applied by the caller.
func NewRouter(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.Files)))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		HealthPage().Render(r.Context(), w)
	})
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.LoginPost)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.HandleFunc("GET /change-password", h.ChangePassPage)
	mux.HandleFunc("POST /change-password", h.ChangePassPost)

	// Admin routes — RequireAdmin middleware applied per-handler
	mux.Handle("GET /admin", RequireAdmin(http.HandlerFunc(h.Admin.Dashboard)))
	mux.Handle("GET /admin/pharmacies/new", RequireAdmin(http.HandlerFunc(h.Admin.NewPharmacy)))
	mux.Handle("POST /admin/pharmacies", RequireAdmin(http.HandlerFunc(h.Admin.CreatePharmacy)))
	mux.Handle("GET /admin/pharmacies/{id}", RequireAdmin(http.HandlerFunc(h.Admin.PharmacyDetail)))
	mux.Handle("POST /admin/pharmacies/{id}", RequireAdmin(http.HandlerFunc(h.Admin.UpdatePharmacy)))
	mux.Handle("GET /admin/pharmacies/{id}/personnel/new", RequireAdmin(http.HandlerFunc(h.Admin.AddPersonnel)))
	mux.Handle("POST /admin/pharmacies/{id}/personnel", RequireAdmin(http.HandlerFunc(h.Admin.CreatePersonnel)))

	// Owner routes — RequireOwner middleware applied per-handler
	mux.Handle("GET /personnel", RequireOwner(http.HandlerFunc(h.Owner.PersonnelList)))

	return mux
}
