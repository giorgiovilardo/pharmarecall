package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
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

func loginTestServer(sm *scs.SessionManager, auth handler.Authenticator) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", handler.HandleLoginPage())
	mux.HandleFunc("POST /login", handler.HandleLoginPost(sm, auth))
	return httptest.NewServer(sm.LoadAndSave(mux))
}

func TestLoginPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := loginTestServer(sm, &stubAuthenticator{})
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
	srv := loginTestServer(sm, auth)
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
	srv := loginTestServer(sm, auth)
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
			srv := loginTestServer(sm, auth)
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
