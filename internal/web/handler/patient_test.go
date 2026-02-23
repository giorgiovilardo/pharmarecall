package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/patient"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

// --- Stubs ---

type stubPatientLister struct {
	pharmacyID int64
	patients   []patient.Summary
	err        error
}

func (s *stubPatientLister) List(_ context.Context, pharmacyID int64) ([]patient.Summary, error) {
	s.pharmacyID = pharmacyID
	return s.patients, s.err
}

type stubPatientCreator struct {
	called bool
	params patient.CreateParams
	result patient.Patient
	err    error
}

func (s *stubPatientCreator) Create(_ context.Context, p patient.CreateParams) (patient.Patient, error) {
	s.called = true
	s.params = p
	return s.result, s.err
}

// --- Test server ---

func patientTestServer(sm *scs.SessionManager, lister handler.PatientLister, creator handler.PatientCreator) *httptest.Server {
	mux := http.NewServeMux()
	if lister != nil {
		mux.Handle("GET /patients", web.RequireAuth(http.HandlerFunc(handler.HandlePatientList(lister))))
	}
	mux.Handle("GET /patients/new", web.RequireAuth(http.HandlerFunc(handler.HandleNewPatientPage())))
	if creator != nil {
		mux.Handle("POST /patients", web.RequireAuth(http.HandlerFunc(handler.HandleCreatePatient(creator))))
	}
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "personnel")
		sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
}

// --- Patient list tests (5.3) ---

func TestPatientListRendersPatients(t *testing.T) {
	lister := &stubPatientLister{
		patients: []patient.Summary{
			{ID: 1, FirstName: "Mario", LastName: "Rossi", Phone: "333-1234567", Email: "mario@example.com", Consensus: true},
			{ID: 2, FirstName: "Anna", LastName: "Verdi", Phone: "333-7654321", Email: "", Consensus: false},
		},
	}

	sm := scs.New()
	srv := patientTestServer(sm, lister, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Mario", "Rossi", "Anna", "Verdi", "333-1234567", "mario@example.com"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}

	if lister.pharmacyID != 7 {
		t.Errorf("pharmacyID passed to lister = %d, want 7", lister.pharmacyID)
	}
}

func TestPatientListEmptyShowsMessage(t *testing.T) {
	lister := &stubPatientLister{patients: nil}

	sm := scs.New()
	srv := patientTestServer(sm, lister, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nessun paziente") {
		t.Error("body missing empty state message")
	}
}

func TestPatientListDatabaseErrorReturns500(t *testing.T) {
	lister := &stubPatientLister{err: errors.New("db down")}

	sm := scs.New()
	srv := patientTestServer(sm, lister, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

// --- New patient form tests (5.4) ---

func TestNewPatientPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := patientTestServer(sm, nil, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients/new")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Nuovo Paziente", "first_name", "last_name", "phone", "email"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

// --- Create patient handler tests (5.5) ---

func TestCreatePatientSuccessRedirects(t *testing.T) {
	stub := &stubPatientCreator{result: patient.Patient{ID: 10}}

	sm := scs.New()
	srv := patientTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"first_name": {"Mario"},
		"last_name":  {"Rossi"},
		"phone":      {"333-1234567"},
		"email":      {"mario@example.com"},
	}
	resp := authenticatedPost(t, srv, "/patients", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/patients" {
		t.Errorf("redirect = %q, want /patients", loc)
	}
	if !stub.called {
		t.Error("create was not called")
	}
	if stub.params.PharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", stub.params.PharmacyID)
	}
	if stub.params.FirstName != "Mario" {
		t.Errorf("firstName = %q, want Mario", stub.params.FirstName)
	}
	if stub.params.Fulfillment != "pickup" {
		t.Errorf("fulfillment = %q, want pickup", stub.params.Fulfillment)
	}
}

func TestCreatePatientMissingNameShowsError(t *testing.T) {
	stub := &stubPatientCreator{}

	sm := scs.New()
	srv := patientTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"first_name": {""},
		"last_name":  {"Rossi"},
		"phone":      {"333-1234567"},
	}
	resp := authenticatedPost(t, srv, "/patients", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}
	if stub.called {
		t.Error("create should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "obbligatori") {
		t.Error("body missing validation error message")
	}
}

func TestCreatePatientMissingContactShowsError(t *testing.T) {
	stub := &stubPatientCreator{}

	sm := scs.New()
	srv := patientTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"first_name": {"Mario"},
		"last_name":  {"Rossi"},
	}
	resp := authenticatedPost(t, srv, "/patients", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}
	if stub.called {
		t.Error("create should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "contatto") {
		t.Error("body missing contact validation error message")
	}
}

func TestCreatePatientShippingWithoutAddressShowsError(t *testing.T) {
	stub := &stubPatientCreator{}

	sm := scs.New()
	srv := patientTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"first_name":  {"Mario"},
		"last_name":   {"Rossi"},
		"phone":       {"333-1234567"},
		"fulfillment": {"shipping"},
	}
	resp := authenticatedPost(t, srv, "/patients", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}
	if stub.called {
		t.Error("create should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "indirizzo") {
		t.Error("body missing address validation error message")
	}
}

func TestCreatePatientServiceErrorReturns500(t *testing.T) {
	stub := &stubPatientCreator{err: errors.New("db down")}

	sm := scs.New()
	srv := patientTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"first_name": {"Mario"},
		"last_name":  {"Rossi"},
		"phone":      {"333-1234567"},
	}
	resp := authenticatedPost(t, srv, "/patients", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
