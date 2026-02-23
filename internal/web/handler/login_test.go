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
	"github.com/giorgiovilardo/pharmarecall/internal/user"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

type stubAuthenticator struct {
	user user.User
	err  error
}

func (s *stubAuthenticator) Authenticate(_ context.Context, _, _ string) (user.User, error) {
	return s.user, s.err
}

type stubPharmacyNameGetter struct {
	pharmacy pharmacy.Pharmacy
	err      error
}

func (s *stubPharmacyNameGetter) Get(_ context.Context, _ int64) (pharmacy.Pharmacy, error) {
	return s.pharmacy, s.err
}

func loginTestServer(sm *scs.SessionManager, auth handler.Authenticator, pharmacies handler.PharmacyNameGetter) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", handler.HandleLoginPage())
	mux.HandleFunc("POST /login", handler.HandleLoginPost(sm, auth, pharmacies))
	return httptest.NewServer(sm.LoadAndSave(mux))
}

func TestLoginPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := loginTestServer(sm, &stubAuthenticator{}, nil)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/login")
	if err != nil {
		t.Fatalf("requesting login page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /login status = %d, want 200", resp.StatusCode)
	}
}

func TestLoginPostValidCredentialsRedirects(t *testing.T) {
	auth := &stubAuthenticator{
		user: user.User{ID: 1, Email: "admin@example.com", Name: "Admin", Role: "admin"},
	}

	sm := scs.New()
	srv := loginTestServer(sm, auth, nil)
	defer srv.Close()

	form := url.Values{"email": {"admin@example.com"}, "password": {"secret123"}}
	resp, err := noFollowClient().Post(srv.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("posting login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("POST /login status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin" {
		t.Errorf("POST /login redirect = %q, want /admin", loc)
	}
}

func TestLoginPostInvalidCredentialsShowsError(t *testing.T) {
	auth := &stubAuthenticator{err: user.ErrInvalidCredentials}

	sm := scs.New()
	srv := loginTestServer(sm, auth, nil)
	defer srv.Close()

	form := url.Values{"email": {"user@example.com"}, "password": {"wrong-password"}}
	resp, err := noFollowClient().Post(srv.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("posting login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("POST /login with bad credentials status = %d, want 200 (re-render form)", resp.StatusCode)
	}
}

func TestLoginPostRedirectsByRole(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		wantLoc string
	}{
		{"admin goes to /admin", "admin", "/admin"},
		{"owner goes to /dashboard", "owner", "/dashboard"},
		{"personnel goes to /dashboard", "personnel", "/dashboard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &stubAuthenticator{
				user: user.User{ID: 1, Email: "user@example.com", Role: tt.role},
			}

			sm := scs.New()
			srv := loginTestServer(sm, auth, nil)
			defer srv.Close()

			form := url.Values{"email": {"user@example.com"}, "password": {"pass"}}
			resp, err := noFollowClient().Post(srv.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
			if err != nil {
				t.Fatalf("posting login: %v", err)
			}
			defer resp.Body.Close()

			if loc := resp.Header.Get("Location"); loc != tt.wantLoc {
				t.Errorf("redirect = %q, want %q", loc, tt.wantLoc)
			}
		})
	}
}

func TestLoginPostStoresUserNameAndPharmacyNameInSession(t *testing.T) {
	auth := &stubAuthenticator{
		user: user.User{ID: 5, Email: "owner@example.com", Name: "Mario Rossi", Role: "owner", PharmacyID: 10},
	}
	pharmacies := &stubPharmacyNameGetter{
		pharmacy: pharmacy.Pharmacy{ID: 10, Name: "Farmacia Centrale"},
	}

	sm := scs.New()

	// We need to verify session contents after login, so add a /check route.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", handler.HandleLoginPage())
	mux.HandleFunc("POST /login", handler.HandleLoginPost(sm, auth, pharmacies))
	mux.HandleFunc("GET /check", func(w http.ResponseWriter, r *http.Request) {
		userName := sm.GetString(r.Context(), "userName")
		pharmacyName := sm.GetString(r.Context(), "pharmacyName")
		w.Write([]byte(userName + "|" + pharmacyName))
	})
	srv := httptest.NewServer(sm.LoadAndSave(mux))
	defer srv.Close()

	client := noFollowClient()
	form := url.Values{"email": {"owner@example.com"}, "password": {"pass"}}
	resp, err := client.Post(srv.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("posting login: %v", err)
	}
	resp.Body.Close()
	cookies := resp.Cookies()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}

	checkResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("checking session: %v", err)
	}
	defer checkResp.Body.Close()

	bodyBytes, _ := io.ReadAll(checkResp.Body)

	if string(bodyBytes) != "Mario Rossi|Farmacia Centrale" {
		t.Errorf("session data = %q, want %q", string(bodyBytes), "Mario Rossi|Farmacia Centrale")
	}
}

func TestLoginPostPharmacyNameErrorIsNonFatal(t *testing.T) {
	auth := &stubAuthenticator{
		user: user.User{ID: 5, Email: "owner@example.com", Name: "Mario Rossi", Role: "owner", PharmacyID: 10},
	}
	pharmacies := &stubPharmacyNameGetter{
		err: errors.New("db connection lost"),
	}

	sm := scs.New()
	srv := loginTestServer(sm, auth, pharmacies)
	defer srv.Close()

	form := url.Values{"email": {"owner@example.com"}, "password": {"pass"}}
	resp, err := noFollowClient().Post(srv.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("posting login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303 (login should succeed despite pharmacy lookup error)", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/dashboard" {
		t.Errorf("redirect = %q, want /dashboard", loc)
	}
}
