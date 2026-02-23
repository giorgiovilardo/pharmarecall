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
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

type stubPersonnelLister struct {
	pharmacyID int64
	members    []pharmacy.PersonnelMember
	err        error
}

func (s *stubPersonnelLister) ListPersonnel(_ context.Context, pharmacyID int64) ([]pharmacy.PersonnelMember, error) {
	s.pharmacyID = pharmacyID
	return s.members, s.err
}

type stubPersonnelCreator struct {
	called bool
	params pharmacy.CreatePersonnelParams
	member pharmacy.PersonnelMember
	err    error
}

func (s *stubPersonnelCreator) CreatePersonnel(_ context.Context, p pharmacy.CreatePersonnelParams) (pharmacy.PersonnelMember, error) {
	s.called = true
	s.params = p
	return s.member, s.err
}

func ownerPersonnelTestServer(sm *scs.SessionManager, lister handler.PersonnelLister, creator handler.PersonnelCreator) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle("GET /personnel", web.RequireOwner(http.HandlerFunc(handler.HandleOwnerPersonnelList(lister))))
	mux.Handle("GET /personnel/new", web.RequireOwner(http.HandlerFunc(handler.HandleOwnerAddPersonnelPage())))
	if creator != nil {
		mux.Handle("POST /personnel", web.RequireOwner(http.HandlerFunc(handler.HandleOwnerCreatePersonnel(creator))))
	}
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "owner")
		sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
}

// --- Personnel list tests ---

func TestOwnerPersonnelListRendersMembers(t *testing.T) {
	lister := &stubPersonnelLister{
		members: []pharmacy.PersonnelMember{
			{ID: 1, Name: "Anna Verdi", Email: "anna@example.com", Role: "personnel"},
			{ID: 2, Name: "Marco Rossi", Email: "marco@example.com", Role: "owner"},
		},
	}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, lister, nil)
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
	lister := &stubPersonnelLister{members: nil}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, lister, nil)
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
	lister := &stubPersonnelLister{err: errors.New("db down")}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, lister, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/personnel")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

// --- Add personnel form tests (4.3) ---

func TestOwnerAddPersonnelPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, nil, nil)
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/personnel/new")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Nuovo Personale", "name", "email", "password"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

// --- Create personnel handler tests (4.4) ---

func TestOwnerCreatePersonnelSuccessRedirects(t *testing.T) {
	stub := &stubPersonnelCreator{member: pharmacy.PersonnelMember{ID: 5}}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"name":     {"Anna Verdi"},
		"email":    {"anna@example.com"},
		"password": {"temppass123"},
	}
	resp := authenticatedPost(t, srv, "/personnel", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/personnel" {
		t.Errorf("redirect = %q, want /personnel", loc)
	}
	if !stub.called {
		t.Error("create function was not called")
	}
	if stub.params.Role != "personnel" {
		t.Errorf("role = %q, want personnel", stub.params.Role)
	}
	if stub.params.PharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", stub.params.PharmacyID)
	}
}

func TestOwnerCreatePersonnelMissingFieldsShowsError(t *testing.T) {
	stub := &stubPersonnelCreator{}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"name":  {"Anna Verdi"},
		"email": {"anna@example.com"},
		// password missing
	}
	resp := authenticatedPost(t, srv, "/personnel", form)
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

func TestOwnerCreatePersonnelDuplicateEmailShowsError(t *testing.T) {
	stub := &stubPersonnelCreator{err: pharmacy.ErrDuplicateEmail}

	sm := scs.New()
	srv := ownerPersonnelTestServer(sm, nil, stub)
	defer srv.Close()

	form := url.Values{
		"name":     {"Anna Verdi"},
		"email":    {"anna@example.com"},
		"password": {"temppass123"},
	}
	resp := authenticatedPost(t, srv, "/personnel", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "già in uso") {
		t.Error("body missing duplicate email error message")
	}
}
