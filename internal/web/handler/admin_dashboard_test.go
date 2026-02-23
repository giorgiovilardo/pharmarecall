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
	"github.com/giorgiovilardo/pharmarecall/internal/pharmacy"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

type stubPharmacyLister struct {
	pharmacies []pharmacy.Summary
	err        error
}

func (s *stubPharmacyLister) List(_ context.Context) ([]pharmacy.Summary, error) {
	return s.pharmacies, s.err
}

func adminDashboardTestServer(sm *scs.SessionManager, lister handler.PharmacyLister) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /admin", web.RequireAdmin(http.HandlerFunc(handler.HandleAdminDashboard(lister))))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
}

func TestAdminDashboardRendersPharmacyList(t *testing.T) {
	lister := &stubPharmacyLister{
		pharmacies: []pharmacy.Summary{
			{ID: 1, Name: "Farmacia Rossi", Address: "Via Roma 1", PersonnelCount: 3},
			{ID: 2, Name: "Farmacia Bianchi", Address: "Via Milano 5", PersonnelCount: 1},
		},
	}

	sm := scs.New()
	srv := adminDashboardTestServer(sm, lister)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Farmacia Rossi", "Via Roma 1", "Farmacia Bianchi", "Via Milano 5"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestAdminDashboardEmptyShowsMessage(t *testing.T) {
	lister := &stubPharmacyLister{pharmacies: nil}

	sm := scs.New()
	srv := adminDashboardTestServer(sm, lister)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nessuna farmacia registrata.") {
		t.Error("body missing empty state message")
	}
}

func TestAdminDashboardDatabaseErrorReturns500(t *testing.T) {
	lister := &stubPharmacyLister{err: errors.New("db down")}

	sm := scs.New()
	srv := adminDashboardTestServer(sm, lister)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/admin")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
