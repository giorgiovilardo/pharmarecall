package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/auth"
	"github.com/giorgiovilardo/pharmarecall/internal/db"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/jackc/pgx/v5"
)

// stubUserGetter returns a fixed user or error for any email lookup.
type stubUserGetter struct {
	user db.User
	err  error
}

func (s *stubUserGetter) GetUserByEmail(_ context.Context, _ string) (db.User, error) {
	return s.user, s.err
}

// emailMatchingUserGetter returns different users based on email.
type emailMatchingUserGetter struct {
	users map[string]db.User
}

func (g *emailMatchingUserGetter) GetUserByEmail(_ context.Context, email string) (db.User, error) {
	u, ok := g.users[email]
	if !ok {
		return db.User{}, pgx.ErrNoRows
	}
	return u, nil
}

// noFollowClient returns an HTTP client that does not follow redirects.
func noFollowClient() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

// loginTestServer builds a test server with just the login POST handler
// wrapped in SCS middleware (required for session operations).
func loginTestServer(sm *scs.SessionManager, users web.UserByEmailGetter) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", web.HandleLoginPage())
	mux.HandleFunc("POST /login", web.HandleLoginPost(sm, users))
	return httptest.NewServer(sm.LoadAndSave(mux))
}

func TestLoginPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := loginTestServer(sm, &stubUserGetter{})
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
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}

	users := &stubUserGetter{
		user: db.User{
			ID:           1,
			Email:        "admin@example.com",
			PasswordHash: hash,
			Name:         "Admin",
			Role:         "admin",
		},
	}

	sm := scs.New()
	srv := loginTestServer(sm, users)
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

func TestLoginPostInvalidPasswordShowsError(t *testing.T) {
	hash, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hashing password: %v", err)
	}

	users := &stubUserGetter{
		user: db.User{
			ID:           1,
			Email:        "user@example.com",
			PasswordHash: hash,
			Role:         "personnel",
		},
	}

	sm := scs.New()
	srv := loginTestServer(sm, users)
	defer srv.Close()

	form := url.Values{"email": {"user@example.com"}, "password": {"wrong-password"}}
	resp, err := noFollowClient().Post(srv.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("posting login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("POST /login with bad password status = %d, want 200 (re-render form)", resp.StatusCode)
	}
}

func TestLoginPostUnknownEmailShowsError(t *testing.T) {
	users := &stubUserGetter{err: pgx.ErrNoRows}

	sm := scs.New()
	srv := loginTestServer(sm, users)
	defer srv.Close()

	form := url.Values{"email": {"nobody@example.com"}, "password": {"whatever"}}
	resp, err := noFollowClient().Post(srv.URL+"/login", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("posting login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("POST /login with unknown email status = %d, want 200 (re-render form)", resp.StatusCode)
	}
}

func TestLoginPostRedirectsByRole(t *testing.T) {
	hash, _ := auth.HashPassword("pass")

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
			users := &emailMatchingUserGetter{users: map[string]db.User{
				"user@example.com": {ID: 1, Email: "user@example.com", PasswordHash: hash, Role: tt.role},
			}}

			sm := scs.New()
			srv := loginTestServer(sm, users)
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
