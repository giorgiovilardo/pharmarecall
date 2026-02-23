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
	PersonnelList   http.HandlerFunc
	AddPersonnel    http.HandlerFunc
	CreatePersonnel http.HandlerFunc
}

// PatientHandlers groups all patient handler funcs (owner + personnel).
type PatientHandlers struct {
	List         http.HandlerFunc
	New          http.HandlerFunc
	Create       http.HandlerFunc
	Detail       http.HandlerFunc
	Update       http.HandlerFunc
	SetConsensus http.HandlerFunc
}

// PrescriptionHandlers groups all prescription handler funcs.
type PrescriptionHandlers struct {
	New          http.HandlerFunc
	Create       http.HandlerFunc
	Edit         http.HandlerFunc
	Update       http.HandlerFunc
	RecordRefill http.HandlerFunc
}

// OrderHandlers groups all order/dashboard handler funcs.
type OrderHandlers struct {
	Dashboard        http.HandlerFunc
	AdvanceStatus    http.HandlerFunc
	PrintDashboard   http.HandlerFunc
	PrintLabel       http.HandlerFunc
	PrintBatchLabels http.HandlerFunc
}

// NotificationHandlers groups all notification handler funcs.
type NotificationHandlers struct {
	List        http.HandlerFunc
	MarkRead    http.HandlerFunc
	MarkAllRead http.HandlerFunc
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
	Patient        PatientHandlers
	Prescription   PrescriptionHandlers
	Order          OrderHandlers
	Notification   NotificationHandlers
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

	// Dashboard — pharmacy staff landing page (order dashboard)
	mux.Handle("GET /dashboard", RequirePharmacyStaff(http.HandlerFunc(h.Order.Dashboard)))
	mux.Handle("GET /dashboard/print", RequirePharmacyStaff(http.HandlerFunc(h.Order.PrintDashboard)))
	mux.Handle("GET /dashboard/labels", RequirePharmacyStaff(http.HandlerFunc(h.Order.PrintBatchLabels)))

	// Order routes — RequirePharmacyStaff middleware
	mux.Handle("POST /orders/{id}/advance", RequirePharmacyStaff(http.HandlerFunc(h.Order.AdvanceStatus)))
	mux.Handle("GET /orders/{id}/label", RequirePharmacyStaff(http.HandlerFunc(h.Order.PrintLabel)))

	// Notification routes — RequirePharmacyStaff middleware
	mux.Handle("GET /notifications", RequirePharmacyStaff(http.HandlerFunc(h.Notification.List)))
	mux.Handle("POST /notifications/{id}/read", RequirePharmacyStaff(http.HandlerFunc(h.Notification.MarkRead)))
	mux.Handle("POST /notifications/read-all", RequirePharmacyStaff(http.HandlerFunc(h.Notification.MarkAllRead)))

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
	mux.Handle("GET /personnel/new", RequireOwner(http.HandlerFunc(h.Owner.AddPersonnel)))
	mux.Handle("POST /personnel", RequireOwner(http.HandlerFunc(h.Owner.CreatePersonnel)))

	// Patient routes — RequirePharmacyStaff middleware (owner + personnel)
	mux.Handle("GET /patients", RequirePharmacyStaff(http.HandlerFunc(h.Patient.List)))
	mux.Handle("GET /patients/new", RequirePharmacyStaff(http.HandlerFunc(h.Patient.New)))
	mux.Handle("POST /patients", RequirePharmacyStaff(http.HandlerFunc(h.Patient.Create)))
	mux.Handle("GET /patients/{id}", RequirePharmacyStaff(http.HandlerFunc(h.Patient.Detail)))
	mux.Handle("POST /patients/{id}", RequirePharmacyStaff(http.HandlerFunc(h.Patient.Update)))
	mux.Handle("POST /patients/{id}/consensus", RequirePharmacyStaff(http.HandlerFunc(h.Patient.SetConsensus)))

	// Prescription routes — RequirePharmacyStaff middleware
	mux.Handle("GET /patients/{id}/prescriptions/new", RequirePharmacyStaff(http.HandlerFunc(h.Prescription.New)))
	mux.Handle("POST /patients/{id}/prescriptions", RequirePharmacyStaff(http.HandlerFunc(h.Prescription.Create)))
	mux.Handle("GET /patients/{id}/prescriptions/{rxid}/edit", RequirePharmacyStaff(http.HandlerFunc(h.Prescription.Edit)))
	mux.Handle("POST /patients/{id}/prescriptions/{rxid}", RequirePharmacyStaff(http.HandlerFunc(h.Prescription.Update)))
	mux.Handle("POST /patients/{id}/prescriptions/{rxid}/refill", RequirePharmacyStaff(http.HandlerFunc(h.Prescription.RecordRefill)))

	return mux
}
