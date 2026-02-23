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
	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

// --- Stubs ---

type stubPharmacyGetter struct {
	pharmacy pharmacy.Pharmacy
	err      error
}

func (s *stubPharmacyGetter) Get(_ context.Context, _ int64) (pharmacy.Pharmacy, error) {
	return s.pharmacy, s.err
}

type stubPharmacyUpdater struct {
	called bool
	params pharmacy.UpdateParams
	err    error
}

func (s *stubPharmacyUpdater) Update(_ context.Context, p pharmacy.UpdateParams) error {
	s.called = true
	s.params = p
	return s.err
}

type stubPharmacyCreator struct {
	called   bool
	pharmacy pharmacy.Pharmacy
	err      error
}

func (s *stubPharmacyCreator) CreateWithOwner(_ context.Context, _ pharmacy.CreateParams) (pharmacy.Pharmacy, error) {
	s.called = true
	return s.pharmacy, s.err
}

// --- Detail test servers ---

func pharmacyDetailTestServer(sm *scs.SessionManager, getter handler.PharmacyGetter, personnel handler.PersonnelLister, updater handler.PharmacyUpdater) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/pharmacies/{id}", handler.HandlePharmacyDetail(getter, personnel))
	if updater != nil {
		mux.HandleFunc("POST /admin/pharmacies/{id}", handler.HandleUpdatePharmacy(getter, personnel, updater))
	}
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(mux))
}

// --- Detail tests ---

func TestPharmacyDetailRendersPharmacyAndPersonnel(t *testing.T) {
	getter := &stubPharmacyGetter{
		pharmacy: pharmacy.Pharmacy{ID: 1, Name: "Farmacia Rossi", Address: "Via Roma 1", Phone: "0123", Email: "f@example.com"},
	}
	personnel := &stubPersonnelLister{
		members: []pharmacy.PersonnelMember{{Name: "Mario Rossi", Email: "mario@example.com", Role: "owner"}},
	}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, getter, personnel, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin/pharmacies/1")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Farmacia Rossi", "Via Roma 1", "Mario Rossi", "mario@example.com"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestPharmacyDetailNotFoundReturns404(t *testing.T) {
	getter := &stubPharmacyGetter{err: pharmacy.ErrNotFound}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, getter, &stubPersonnelLister{}, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin/pharmacies/999")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPharmacyDetailInvalidIDReturns404(t *testing.T) {
	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, &stubPharmacyGetter{}, &stubPersonnelLister{}, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin/pharmacies/abc")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPharmacyDetailEmptyPersonnelShowsMessage(t *testing.T) {
	getter := &stubPharmacyGetter{
		pharmacy: pharmacy.Pharmacy{ID: 1, Name: "Farmacia Rossi"},
	}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, getter, &stubPersonnelLister{}, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin/pharmacies/1")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nessun personale.") {
		t.Error("body missing empty personnel message")
	}
}

// --- Update tests ---

func TestUpdatePharmacySuccessRedirects(t *testing.T) {
	getter := &stubPharmacyGetter{pharmacy: pharmacy.Pharmacy{ID: 1}}
	updater := &stubPharmacyUpdater{}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, getter, &stubPersonnelLister{}, updater)
	defer srv.Close()

	form := url.Values{
		"name":    {"Farmacia Nuova"},
		"address": {"Via Milano 10"},
		"phone":   {"999"},
		"email":   {"new@example.com"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies/1", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/pharmacies/1" {
		t.Errorf("redirect = %q, want /admin/pharmacies/1", loc)
	}
	if !updater.called {
		t.Error("update function was not called")
	}
	if updater.params.Name != "Farmacia Nuova" {
		t.Errorf("update name = %q, want Farmacia Nuova", updater.params.Name)
	}
}

func TestUpdatePharmacyMissingFieldsShowsError(t *testing.T) {
	getter := &stubPharmacyGetter{pharmacy: pharmacy.Pharmacy{ID: 1}}
	updater := &stubPharmacyUpdater{}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, getter, &stubPersonnelLister{}, updater)
	defer srv.Close()

	form := url.Values{
		"name": {"Farmacia Nuova"},
		// address missing
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies/1", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render)", resp.StatusCode)
	}
	if updater.called {
		t.Error("update function should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "obbligatori") {
		t.Error("body missing validation error message")
	}
}

// --- Create pharmacy tests ---

func createPharmacyTestServer(sm *scs.SessionManager, creator handler.PharmacyCreatorWithOwner) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/pharmacies/new", handler.HandleNewPharmacyPage())
	mux.HandleFunc("POST /admin/pharmacies", handler.HandleCreatePharmacy(creator))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(mux))
}

func TestNewPharmacyPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := createPharmacyTestServer(sm, nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/admin/pharmacies/new")
	if err != nil {
		t.Fatalf("requesting new pharmacy page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nuova Farmacia") {
		t.Error("page missing title")
	}
}

func TestCreatePharmacyMissingFieldsShowsError(t *testing.T) {
	sm := scs.New()
	srv := createPharmacyTestServer(sm, nil)
	defer srv.Close()

	form := url.Values{
		"pharmacy_name": {"Farmacia Test"},
		// address missing
		"owner_name":     {"Mario Rossi"},
		"owner_email":    {"mario@example.com"},
		"owner_password": {"secret123"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "obbligatori") {
		t.Error("body missing validation error message")
	}
}

func TestCreatePharmacySuccessRedirects(t *testing.T) {
	creator := &stubPharmacyCreator{pharmacy: pharmacy.Pharmacy{ID: 1, Name: "Farmacia Test"}}

	sm := scs.New()
	srv := createPharmacyTestServer(sm, creator)
	defer srv.Close()

	form := url.Values{
		"pharmacy_name":  {"Farmacia Test"},
		"address":        {"Via Roma 1"},
		"phone":          {"0123456789"},
		"email":          {"farmacia@example.com"},
		"owner_name":     {"Mario Rossi"},
		"owner_email":    {"mario@example.com"},
		"owner_password": {"secret123"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin" {
		t.Errorf("redirect = %q, want /admin", loc)
	}
}

func TestCreatePharmacyDuplicateEmailShowsError(t *testing.T) {
	creator := &stubPharmacyCreator{err: pharmacy.ErrDuplicateEmail}

	sm := scs.New()
	srv := createPharmacyTestServer(sm, creator)
	defer srv.Close()

	form := url.Values{
		"pharmacy_name":  {"Farmacia Test"},
		"address":        {"Via Roma 1"},
		"owner_name":     {"Mario Rossi"},
		"owner_email":    {"mario@example.com"},
		"owner_password": {"secret123"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "già in uso") {
		t.Error("body missing duplicate email error message")
	}
}

func TestCreatePharmacyDBErrorReturns500(t *testing.T) {
	creator := &stubPharmacyCreator{err: errors.New("connection refused")}

	sm := scs.New()
	srv := createPharmacyTestServer(sm, creator)
	defer srv.Close()

	form := url.Values{
		"pharmacy_name":  {"Farmacia Test"},
		"address":        {"Via Roma 1"},
		"owner_name":     {"Mario Rossi"},
		"owner_email":    {"mario@example.com"},
		"owner_password": {"secret123"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
