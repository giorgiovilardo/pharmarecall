package web_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/jackc/pgx/v5"
)

type stubPharmacyDetailReader struct {
	pharmacy  db.Pharmacy
	getErr    error
	personnel []db.User
	listErr   error
}

func (s *stubPharmacyDetailReader) GetPharmacyByID(_ context.Context, _ int64) (db.Pharmacy, error) {
	return s.pharmacy, s.getErr
}

func (s *stubPharmacyDetailReader) ListUsersByPharmacy(_ context.Context, _ int64) ([]db.User, error) {
	return s.personnel, s.listErr
}

type stubUpdatePharmacy struct {
	called bool
	params db.UpdatePharmacyParams
	err    error
}

func (s *stubUpdatePharmacy) fn() web.UpdatePharmacyFunc {
	return func(_ context.Context, arg db.UpdatePharmacyParams) error {
		s.called = true
		s.params = arg
		return s.err
	}
}

func pharmacyDetailTestServer(sm *scs.SessionManager, store web.PharmacyDetailReader, updateFn web.UpdatePharmacyFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/pharmacies/{id}", web.HandlePharmacyDetail(store))
	mux.HandleFunc("POST /admin/pharmacies/{id}", web.HandleUpdatePharmacy(store, updateFn))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(mux))
}

func TestPharmacyDetailRendersPharmacyAndPersonnel(t *testing.T) {
	store := &stubPharmacyDetailReader{
		pharmacy:  db.Pharmacy{ID: 1, Name: "Farmacia Rossi", Address: "Via Roma 1", Phone: "0123", Email: "f@example.com"},
		personnel: []db.User{{Name: "Mario Rossi", Email: "mario@example.com", Role: "owner"}},
	}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, store, nil)
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
	store := &stubPharmacyDetailReader{getErr: pgx.ErrNoRows}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, store, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin/pharmacies/999")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPharmacyDetailInvalidIDReturns404(t *testing.T) {
	store := &stubPharmacyDetailReader{}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, store, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin/pharmacies/abc")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPharmacyDetailEmptyPersonnelShowsMessage(t *testing.T) {
	store := &stubPharmacyDetailReader{
		pharmacy:  db.Pharmacy{ID: 1, Name: "Farmacia Rossi"},
		personnel: nil,
	}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, store, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin/pharmacies/1")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nessun personale.") {
		t.Error("body missing empty personnel message")
	}
}

func TestUpdatePharmacySuccessRedirects(t *testing.T) {
	store := &stubPharmacyDetailReader{pharmacy: db.Pharmacy{ID: 1}}
	update := &stubUpdatePharmacy{}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, store, update.fn())
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
	if !update.called {
		t.Error("update function was not called")
	}
	if update.params.Name != "Farmacia Nuova" {
		t.Errorf("update name = %q, want Farmacia Nuova", update.params.Name)
	}
}

func TestUpdatePharmacyMissingFieldsShowsError(t *testing.T) {
	store := &stubPharmacyDetailReader{pharmacy: db.Pharmacy{ID: 1}}
	update := &stubUpdatePharmacy{}

	sm := scs.New()
	srv := pharmacyDetailTestServer(sm, store, update.fn())
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
	if update.called {
		t.Error("update function should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "obbligatori") {
		t.Error("body missing validation error message")
	}
}
