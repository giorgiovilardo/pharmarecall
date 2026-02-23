package web_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

type stubPharmacyLister struct {
	pharmacies []db.ListPharmaciesRow
	err        error
}

func (s *stubPharmacyLister) ListPharmacies(_ context.Context) ([]db.ListPharmaciesRow, error) {
	return s.pharmacies, s.err
}

func adminDashboardTestServer(sm *scs.SessionManager, lister web.PharmacyLister) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /admin", web.RequireAdmin(http.HandlerFunc(web.HandleAdminDashboard(lister))))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
}

func authenticatedGet(t *testing.T, srv *httptest.Server, path string) *http.Response {
	t.Helper()
	client := noFollowClient()

	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting %s: %v", path, err)
	}
	return resp
}

func TestAdminDashboardRendersPharmacyList(t *testing.T) {
	lister := &stubPharmacyLister{
		pharmacies: []db.ListPharmaciesRow{
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
