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
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/patient"
	"github.com/giorgiovilardo/pharmarecall/internal/prescription"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

// --- Prescription stubs ---

type stubRxCreator struct {
	called bool
	params prescription.CreateParams
	result prescription.Prescription
	err    error
}

func (s *stubRxCreator) Create(_ context.Context, p prescription.CreateParams) (prescription.Prescription, error) {
	s.called = true
	s.params = p
	return s.result, s.err
}

type stubRxGetter struct {
	rx  prescription.Prescription
	err error
}

func (s *stubRxGetter) Get(_ context.Context, _ int64) (prescription.Prescription, error) {
	return s.rx, s.err
}

type stubRxUpdater struct {
	called bool
	params prescription.UpdateParams
	err    error
}

func (s *stubRxUpdater) Update(_ context.Context, p prescription.UpdateParams) error {
	s.called = true
	s.params = p
	return s.err
}

type stubRxRefiller struct {
	called         bool
	prescriptionID int64
	newStartDate   time.Time
	err            error
}

func (s *stubRxRefiller) RecordRefill(_ context.Context, prescriptionID int64, newStartDate time.Time) error {
	s.called = true
	s.prescriptionID = prescriptionID
	s.newStartDate = newStartDate
	return s.err
}

// --- Prescription test server ---

type rxTestDeps struct {
	sm            *scs.SessionManager
	patientGetter handler.PatientGetter
	rxCreator     handler.PrescriptionCreator
	rxGetter      handler.PrescriptionGetter
	rxUpdater     handler.PrescriptionUpdater
	rxRefiller    handler.PrescriptionRefiller
}

func rxTestServer(d rxTestDeps) *httptest.Server {
	mux := http.NewServeMux()
	if d.patientGetter != nil {
		mux.Handle("GET /patients/{id}/prescriptions/new", web.RequireAuth(http.HandlerFunc(handler.HandleNewPrescriptionPage(d.patientGetter))))
	}
	if d.rxCreator != nil && d.patientGetter != nil {
		mux.Handle("POST /patients/{id}/prescriptions", web.RequireAuth(http.HandlerFunc(handler.HandleCreatePrescription(d.rxCreator, d.patientGetter))))
	}
	if d.rxGetter != nil && d.patientGetter != nil {
		mux.Handle("GET /patients/{id}/prescriptions/{rxid}/edit", web.RequireAuth(http.HandlerFunc(handler.HandlePrescriptionEditPage(d.rxGetter, d.patientGetter))))
	}
	if d.rxUpdater != nil && d.rxGetter != nil && d.patientGetter != nil {
		mux.Handle("POST /patients/{id}/prescriptions/{rxid}", web.RequireAuth(http.HandlerFunc(handler.HandleUpdatePrescription(d.rxUpdater, d.rxGetter, d.patientGetter))))
	}
	if d.rxRefiller != nil {
		mux.Handle("POST /patients/{id}/prescriptions/{rxid}/refill", web.RequireAuth(http.HandlerFunc(handler.HandleRecordRefill(d.rxRefiller))))
	}
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		d.sm.Put(r.Context(), "userID", int64(1))
		d.sm.Put(r.Context(), "role", "personnel")
		d.sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(d.sm.LoadAndSave(web.LoadUser(d.sm)(mux)))
}

// --- New prescription form tests (6.6) ---

func TestNewPrescriptionPageRendersForm(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{
		ID: 10, FirstName: "Mario", LastName: "Rossi", Consensus: true,
	}}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: getter})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients/10/prescriptions/new")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Nuova Prescrizione", "Mario", "Rossi", "medication_name", "units_per_box", "daily_consumption", "box_start_date"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestNewPrescriptionPageRedirectsWithoutConsensus(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{
		ID: 10, FirstName: "Mario", LastName: "Rossi", Consensus: false,
	}}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: getter})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients/10/prescriptions/new")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/patients/10" {
		t.Errorf("redirect = %q, want /patients/10", loc)
	}
}

// --- Create prescription handler tests (6.7) ---

func TestCreatePrescriptionSuccessRedirects(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10, Consensus: true}}
	creator := &stubRxCreator{result: prescription.Prescription{ID: 1}}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: getter, rxCreator: creator})
	defer srv.Close()

	form := url.Values{
		"medication_name":   {"Tachipirina"},
		"units_per_box":     {"30"},
		"daily_consumption": {"1"},
		"box_start_date":    {"2026-01-01"},
	}
	resp := authenticatedPost(t, srv, "/patients/10/prescriptions", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/patients/10" {
		t.Errorf("redirect = %q, want /patients/10", loc)
	}
	if !creator.called {
		t.Error("Create was not called")
	}
	if creator.params.MedicationName != "Tachipirina" {
		t.Errorf("MedicationName = %q, want Tachipirina", creator.params.MedicationName)
	}
	if creator.params.UnitsPerBox != 30 {
		t.Errorf("UnitsPerBox = %d, want 30", creator.params.UnitsPerBox)
	}
}

func TestCreatePrescriptionMissingNameShowsError(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10, Consensus: true}}
	creator := &stubRxCreator{}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: getter, rxCreator: creator})
	defer srv.Close()

	form := url.Values{
		"medication_name":   {""},
		"units_per_box":     {"30"},
		"daily_consumption": {"1"},
		"box_start_date":    {"2026-01-01"},
	}
	resp := authenticatedPost(t, srv, "/patients/10/prescriptions", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}
	if creator.called {
		t.Error("Create should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "farmaco") {
		t.Error("body missing validation error about medication name")
	}
}

func TestCreatePrescriptionNoConsensusShowsError(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10, Consensus: true}}
	creator := &stubRxCreator{err: prescription.ErrNoConsensus}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: getter, rxCreator: creator})
	defer srv.Close()

	form := url.Values{
		"medication_name":   {"Tachipirina"},
		"units_per_box":     {"30"},
		"daily_consumption": {"1"},
		"box_start_date":    {"2026-01-01"},
	}
	resp := authenticatedPost(t, srv, "/patients/10/prescriptions", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "consenso") {
		t.Error("body missing consensus error message")
	}
}

func TestCreatePrescriptionServiceErrorReturns500(t *testing.T) {
	getter := &stubPatientGetter{patient: patient.Patient{ID: 10, Consensus: true}}
	creator := &stubRxCreator{err: errors.New("db down")}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: getter, rxCreator: creator})
	defer srv.Close()

	form := url.Values{
		"medication_name":   {"Tachipirina"},
		"units_per_box":     {"30"},
		"daily_consumption": {"1"},
		"box_start_date":    {"2026-01-01"},
	}
	resp := authenticatedPost(t, srv, "/patients/10/prescriptions", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

// --- Prescription edit page tests (6.8) ---

func TestPrescriptionEditPageRendersForm(t *testing.T) {
	pGetter := &stubPatientGetter{patient: patient.Patient{ID: 10, FirstName: "Mario", LastName: "Rossi"}}
	rxGetter := &stubRxGetter{rx: prescription.Prescription{
		ID: 5, PatientID: 10, MedicationName: "Tachipirina",
		UnitsPerBox: 30, DailyConsumption: 1, BoxStartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: pGetter, rxGetter: rxGetter})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients/10/prescriptions/5/edit")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Modifica Prescrizione", "Tachipirina", "30", "2026-01-01"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestPrescriptionEditPageNotFoundReturns404(t *testing.T) {
	pGetter := &stubPatientGetter{patient: patient.Patient{ID: 10}}
	rxGetter := &stubRxGetter{err: prescription.ErrNotFound}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: pGetter, rxGetter: rxGetter})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients/10/prescriptions/999/edit")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// --- Update prescription handler tests (6.8) ---

func TestUpdatePrescriptionSuccessRedirects(t *testing.T) {
	pGetter := &stubPatientGetter{patient: patient.Patient{ID: 10}}
	rxGetter := &stubRxGetter{rx: prescription.Prescription{ID: 5, PatientID: 10, MedicationName: "Tachipirina"}}
	rxUpdater := &stubRxUpdater{}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: pGetter, rxGetter: rxGetter, rxUpdater: rxUpdater})
	defer srv.Close()

	form := url.Values{
		"medication_name":   {"Tachipirina 500"},
		"units_per_box":     {"60"},
		"daily_consumption": {"2"},
		"box_start_date":    {"2026-02-01"},
	}
	resp := authenticatedPost(t, srv, "/patients/10/prescriptions/5", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/patients/10" {
		t.Errorf("redirect = %q, want /patients/10", loc)
	}
	if !rxUpdater.called {
		t.Error("Update was not called")
	}
	if rxUpdater.params.MedicationName != "Tachipirina 500" {
		t.Errorf("MedicationName = %q, want Tachipirina 500", rxUpdater.params.MedicationName)
	}
}

func TestUpdatePrescriptionMissingNameShowsError(t *testing.T) {
	pGetter := &stubPatientGetter{patient: patient.Patient{ID: 10}}
	rxGetter := &stubRxGetter{rx: prescription.Prescription{ID: 5, MedicationName: "Tachipirina"}}
	rxUpdater := &stubRxUpdater{}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, patientGetter: pGetter, rxGetter: rxGetter, rxUpdater: rxUpdater})
	defer srv.Close()

	form := url.Values{
		"medication_name":   {""},
		"units_per_box":     {"30"},
		"daily_consumption": {"1"},
		"box_start_date":    {"2026-01-01"},
	}
	resp := authenticatedPost(t, srv, "/patients/10/prescriptions/5", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}
	if rxUpdater.called {
		t.Error("Update should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "farmaco") {
		t.Error("body missing validation error about medication name")
	}
}

// --- Refill handler tests (6.9) ---

func TestRecordRefillSuccessRedirects(t *testing.T) {
	refiller := &stubRxRefiller{}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, rxRefiller: refiller})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/patients/10/prescriptions/5/refill", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/patients/10" {
		t.Errorf("redirect = %q, want /patients/10", loc)
	}
	if !refiller.called {
		t.Error("RecordRefill was not called")
	}
	if refiller.prescriptionID != 5 {
		t.Errorf("PrescriptionID = %d, want 5", refiller.prescriptionID)
	}
}

func TestRecordRefillErrorReturns500(t *testing.T) {
	refiller := &stubRxRefiller{err: errors.New("db down")}

	sm := scs.New()
	srv := rxTestServer(rxTestDeps{sm: sm, rxRefiller: refiller})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/patients/10/prescriptions/5/refill", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
