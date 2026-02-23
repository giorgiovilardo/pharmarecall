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

type stubOwnerPersonnelLister struct {
	pharmacyID int64
	members    []pharmacy.PersonnelMember
	err        error
}

func (s *stubOwnerPersonnelLister) ListPersonnel(_ context.Context, pharmacyID int64) ([]pharmacy.PersonnelMember, error) {
	s.pharmacyID = pharmacyID
	return s.members, s.err
}

func ownerPersonnelTestServer(sm *scs.SessionManager, lister handler.PersonnelLister) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /personnel", web.RequireOwner(http.HandlerFunc(handler.HandleOwnerPersonnelList(lister))))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "owner")
		sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
}

func TestOwnerPersonnelListRendersMembers(t *testing.T) {
	lister := &stubOwnerPersonnelLister{
		members: []pharmacy.PersonnelMember{
			{ID: 1, Name: "Anna Verdi", Email: "anna@example.com", Role: "personnel"},
			{ID: 2, Name: "Marco Rossi", Email: "marco@example.com", Role: "owner"},
		},
	}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, lister)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/personnel")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Anna Verdi", "anna@example.com", "Marco Rossi", "marco@example.com"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}

	if lister.pharmacyID != 7 {
		t.Errorf("pharmacyID passed to lister = %d, want 7", lister.pharmacyID)
	}
}

func TestOwnerPersonnelListEmptyShowsMessage(t *testing.T) {
	lister := &stubOwnerPersonnelLister{members: nil}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, lister)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/personnel")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nessun personale") {
		t.Error("body missing empty state message")
	}
}

func TestOwnerPersonnelListDatabaseErrorReturns500(t *testing.T) {
	lister := &stubOwnerPersonnelLister{err: errors.New("db down")}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, lister)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/personnel")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
