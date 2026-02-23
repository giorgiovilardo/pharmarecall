package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
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

// --- Test server ---

func patientTestServer(sm *scs.SessionManager, lister handler.PatientLister) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /patients", web.RequireAuth(http.HandlerFunc(handler.HandlePatientList(lister))))
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
	srv := patientTestServer(sm, lister)
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
	srv := patientTestServer(sm, lister)
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
	srv := patientTestServer(sm, lister)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/patients")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
