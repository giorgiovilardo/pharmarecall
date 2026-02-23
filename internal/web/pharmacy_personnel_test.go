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
	"github.com/jackc/pgx/v5/pgconn"
)

type stubCreatePersonnel struct {
	called bool
	params db.CreateUserParams
	user   db.User
	err    error
}

func (s *stubCreatePersonnel) fn() web.CreatePersonnelFunc {
	return func(_ context.Context, arg db.CreateUserParams) (db.User, error) {
		s.called = true
		s.params = arg
		return s.user, s.err
	}
}

func personnelTestServer(sm *scs.SessionManager, createFn web.CreatePersonnelFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/pharmacies/{id}/personnel/new", web.HandleAddPersonnelPage())
	mux.HandleFunc("POST /admin/pharmacies/{id}/personnel", web.HandleCreatePersonnel(createFn))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(mux))
}

func TestAddPersonnelPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := personnelTestServer(sm, nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/admin/pharmacies/1/personnel/new")
	if err != nil {
		t.Fatalf("requesting add personnel page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nuovo Personale") {
		t.Error("page missing title")
	}
}

func TestCreatePersonnelSuccessRedirects(t *testing.T) {
	stub := &stubCreatePersonnel{user: db.User{ID: 5}}

	sm := scs.New()
	srv := personnelTestServer(sm, stub.fn())
	defer srv.Close()

	form := url.Values{
		"name":     {"Anna Verdi"},
		"email":    {"anna@example.com"},
		"password": {"temppass123"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies/1/personnel", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/pharmacies/1" {
		t.Errorf("redirect = %q, want /admin/pharmacies/1", loc)
	}
	if !stub.called {
		t.Error("create function was not called")
	}
	if stub.params.Role != "personnel" {
		t.Errorf("role = %q, want personnel", stub.params.Role)
	}
	if !stub.params.PharmacyID.Valid || stub.params.PharmacyID.Int64 != 1 {
		t.Errorf("pharmacyID = %+v, want {Int64:1, Valid:true}", stub.params.PharmacyID)
	}
}

func TestCreatePersonnelWithOwnerCheckbox(t *testing.T) {
	stub := &stubCreatePersonnel{user: db.User{ID: 6}}

	sm := scs.New()
	srv := personnelTestServer(sm, stub.fn())
	defer srv.Close()

	form := url.Values{
		"name":     {"Luigi Bianchi"},
		"email":    {"luigi@example.com"},
		"password": {"temppass123"},
		"owner":    {"true"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies/1/personnel", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if stub.params.Role != "owner" {
		t.Errorf("role = %q, want owner", stub.params.Role)
	}
}

func TestCreatePersonnelMissingFieldsShowsError(t *testing.T) {
	stub := &stubCreatePersonnel{}

	sm := scs.New()
	srv := personnelTestServer(sm, stub.fn())
	defer srv.Close()

	form := url.Values{
		"name":  {"Anna Verdi"},
		"email": {"anna@example.com"},
		// password missing
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies/1/personnel", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}
	if stub.called {
		t.Error("create function should not have been called")
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "obbligatori") {
		t.Error("body missing validation error message")
	}
}

func TestCreatePersonnelDuplicateEmailShowsError(t *testing.T) {
	stub := &stubCreatePersonnel{err: &pgconn.PgError{Code: "23505"}}

	sm := scs.New()
	srv := personnelTestServer(sm, stub.fn())
	defer srv.Close()

	form := url.Values{
		"name":     {"Anna Verdi"},
		"email":    {"anna@example.com"},
		"password": {"temppass123"},
	}
	resp := authenticatedPost(t, srv, "/admin/pharmacies/1/personnel", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "già in uso") {
		t.Error("body missing duplicate email error message")
	}
}
