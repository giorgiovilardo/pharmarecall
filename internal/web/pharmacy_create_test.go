package web_test

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
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/jackc/pgx/v5/pgconn"
)

func createPharmacyTestServer(sm *scs.SessionManager, createFn web.CreatePharmacyWithOwnerFunc) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/pharmacies/new", web.HandleNewPharmacyPage())
	mux.HandleFunc("POST /admin/pharmacies", web.HandleCreatePharmacy(createFn))
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
	createFn := func(_ context.Context, p db.CreatePharmacyParams, owner db.CreateUserParams) (db.Pharmacy, error) {
		return db.Pharmacy{ID: 1, Name: p.Name}, nil
	}

	sm := scs.New()
	srv := createPharmacyTestServer(sm, createFn)
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
	createFn := func(_ context.Context, _ db.CreatePharmacyParams, _ db.CreateUserParams) (db.Pharmacy, error) {
		return db.Pharmacy{}, &pgconn.PgError{Code: "23505"}
	}

	sm := scs.New()
	srv := createPharmacyTestServer(sm, createFn)
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
	createFn := func(_ context.Context, _ db.CreatePharmacyParams, _ db.CreateUserParams) (db.Pharmacy, error) {
		return db.Pharmacy{}, errors.New("connection refused")
	}

	sm := scs.New()
	srv := createPharmacyTestServer(sm, createFn)
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
