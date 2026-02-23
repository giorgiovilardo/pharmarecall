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
	"github.com/giorgiovilardo/pharmarecall/internal/prescription"
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

type stubPatientGetter struct {
	id      int64
	patient patient.Patient
	err     error
}

func (s *stubPatientGetter) Get(_ context.Context, id int64) (patient.Patient, error) {
	s.id = id
	return s.patient, s.err
}

type stubPatientUpdater struct {
	called bool
	params patient.UpdateParams
	err    error
}

func (s *stubPatientUpdater) Update(_ context.Context, p patient.UpdateParams) error {
	s.called = true
	s.params = p
	return s.err
}

type stubConsensusRecorder struct {
	called bool
	id     int64
	err    error
}

func (s *stubConsensusRecorder) SetConsensus(_ context.Context, id int64) error {
	s.called = true
	s.id = id
	return s.err
}

type stubPrescriptionLister struct {
	prescriptions []prescription.Prescription
	err           error
}

func (s *stubPrescriptionLister) ListByPatient(_ context.Context, _ int64) ([]prescription.Prescription, error) {
	return s.prescriptions, s.err
}

// --- Test server ---

type patientTestDeps struct {
	sm        *scs.SessionManager
	lister    handler.PatientLister
	creator   handler.PatientCreator
	getter    handler.PatientGetter
	updater   handler.PatientUpdater
	consensus handler.PatientConsensusRecorder
	rxLister  handler.PrescriptionLister
}

func patientTestServer(sm *scs.SessionManager, lister handler.PatientLister, creator handler.PatientCreator) *httptest.Server {
	return patientTestServerFull(patientTestDeps{sm: sm, lister: lister, creator: creator})
}

func patientTestServerFull(d patientTestDeps) *httptest.Server {
	if d.rxLister == nil {
		d.rxLister = &stubPrescriptionLister{}
	}
	mux := http.NewServeMux()
	if d.lister != nil {
		mux.Handle("GET /patients", web.RequireAuth(http.HandlerFunc(handler.HandlePatientList(d.lister))))
	}
	mux.Handle("GET /patients/new", web.RequireAuth(http.HandlerFunc(handler.HandleNewPatientPage())))
	if d.creator != nil {
		mux.Handle("POST /patients", web.RequireAuth(http.HandlerFunc(handler.HandleCreatePatient(d.creator))))
	}
	if d.getter != nil {
		mux.Handle("GET /patients/{id}", web.RequireAuth(http.HandlerFunc(handler.HandlePatientDetail(d.getter, d.rxLister))))
	}
	if d.getter != nil && d.updater != nil {
		mux.Handle("POST /patients/{id}", web.RequireAuth(http.HandlerFunc(handler.HandleUpdatePatient(d.getter, d.updater, d.rxLister))))
	}
	if d.consensus != nil {
		mux.Handle("POST /patients/{id}/consensus", web.RequireAuth(http.HandlerFunc(handler.HandleSetConsensus(d.consensus))))
	}
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		d.sm.Put(r.Context(), "userID", int64(1))
		d.sm.Put(r.Context(), "role", "personnel")
		d.sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(d.sm.LoadAndSave(web.LoadUser(d.sm)(mux)))
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
}

func TestCreatePatientMissingNameShowsError(t *testing.T) {
	stub := &stubPatientCreator{err: patient.ErrNameRequired}

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

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "obbligatori") {
		t.Error("body missing validation error message")
	}
}

func TestCreatePatientMissingContactShowsError(t *testing.T) {
	stub := &stubPatientCreator{err: patient.ErrContactRequired}

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

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "contatto") {
		t.Error("body missing contact validation error message")
	}
}

func TestCreatePatientShippingWithoutAddressShowsError(t *testing.T) {
	stub := &stubPatientCreator{err: patient.ErrDeliveryAddrRequired}

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

// --- Patient detail tests (5.6) ---

func TestPatientDetailRendersPatient(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{
		ID: 10, PharmacyID: 7, FirstName: "Mario", LastName: "Rossi",
		Phone: "333-1234567", Email: "mario@example.com",
		DeliveryAddress: "Via Roma 1", Fulfillment: "pickup", Notes: "Nota test",
	}}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, getter: getter})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients/10")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Mario", "Rossi", "333-1234567", "mario@example.com", "Via Roma 1", "Nota test"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestPatientDetailNotFoundReturns404(t *testing.T) {
	getter := &stubPatientGetter{err: patient.ErrNotFound}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, getter: getter})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients/999")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// --- Update patient handler tests (5.7) ---

func TestUpdatePatientSuccessRedirects(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10}}
	updater := &stubPatientUpdater{}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, getter: getter, updater: updater})
	defer srv.Close()

	form := url.Values{
		"first_name":  {"Mario"},
		"last_name":   {"Bianchi"},
		"phone":       {"333-9999999"},
		"email":       {"mario@example.com"},
		"fulfillment": {"pickup"},
	}
	resp := authenticatedPost(t, srv, "/patients/10", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/patients/10" {
		t.Errorf("redirect = %q, want /patients/10", loc)
	}
	if !updater.called {
		t.Error("update was not called")
	}
	if updater.params.LastName != "Bianchi" {
		t.Errorf("lastName = %q, want Bianchi", updater.params.LastName)
	}
}

func TestUpdatePatientMissingNameShowsError(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10, FirstName: "Mario", LastName: "Rossi"}}
	updater := &stubPatientUpdater{err: patient.ErrNameRequired}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, getter: getter, updater: updater})
	defer srv.Close()

	form := url.Values{
		"first_name": {""},
		"last_name":  {"Rossi"},
		"phone":      {"333-1234567"},
	}
	resp := authenticatedPost(t, srv, "/patients/10", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "obbligatori") {
		t.Error("body missing validation error message")
	}
}

func TestUpdatePatientMissingContactShowsError(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10, FirstName: "Mario", LastName: "Rossi"}}
	updater := &stubPatientUpdater{err: patient.ErrContactRequired}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, getter: getter, updater: updater})
	defer srv.Close()

	form := url.Values{
		"first_name": {"Mario"},
		"last_name":  {"Rossi"},
	}
	resp := authenticatedPost(t, srv, "/patients/10", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "contatto") {
		t.Error("body missing contact validation error message")
	}
}

func TestUpdatePatientShippingWithoutAddressShowsError(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10, FirstName: "Mario", LastName: "Rossi"}}
	updater := &stubPatientUpdater{err: patient.ErrDeliveryAddrRequired}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, getter: getter, updater: updater})
	defer srv.Close()

	form := url.Values{
		"first_name":  {"Mario"},
		"last_name":   {"Rossi"},
		"phone":       {"333-1234567"},
		"fulfillment": {"shipping"},
	}
	resp := authenticatedPost(t, srv, "/patients/10", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "indirizzo") {
		t.Error("body missing address validation error message")
	}
}

// --- Consensus recording tests (5.8) ---

func TestSetConsensusSuccessRedirects(t *testing.T) {
	recorder := &stubConsensusRecorder{}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, consensus: recorder})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/patients/10/consensus", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/patients/10" {
		t.Errorf("redirect = %q, want /patients/10", loc)
	}
	if !recorder.called {
		t.Error("SetConsensus was not called")
	}
	if recorder.id != 10 {
		t.Errorf("id = %d, want 10", recorder.id)
	}
}

func TestSetConsensusErrorReturns500(t *testing.T) {
	recorder := &stubConsensusRecorder{err: errors.New("db down")}

	sm := scs.New()
	srv := patientTestServerFull(patientTestDeps{sm: sm, consensus: recorder})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/patients/10/consensus", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
