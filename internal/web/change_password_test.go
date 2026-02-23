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
	"golang.org/x/crypto/bcrypt"
)

type stubPasswordChanger struct {
	user         db.User
	getErr       error
	updateParams db.UpdateUserPasswordParams
	updateCalled bool
	updateErr    error
}

func (s *stubPasswordChanger) GetUserByID(_ context.Context, id int64) (db.User, error) {
	return s.user, s.getErr
}

func (s *stubPasswordChanger) UpdateUserPassword(_ context.Context, arg db.UpdateUserPasswordParams) error {
	s.updateCalled = true
	s.updateParams = arg
	return s.updateErr
}

// changePasswordTestServer builds a test server with session setup + change password routes.
func changePasswordTestServer(sm *scs.SessionManager, users web.PasswordChanger) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /change-password", web.HandleChangePasswordPage())
	mux.HandleFunc("POST /change-password", web.HandleChangePasswordPost(sm, users))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(mux))
}

// authenticatedPost creates a session then posts to the given URL with form data.
func authenticatedPost(t *testing.T, srv *httptest.Server, path string, form url.Values) *http.Response {
	t.Helper()
	client := noFollowClient()

	// Set up session
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()
	cookies := setupResp.Cookies()

	// POST with session cookie
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("posting: %v", err)
	}
	return resp
}

func TestChangePasswordPageRendersForm(t *testing.T) {
	sm := scs.New()
	srv := changePasswordTestServer(sm, &stubPasswordChanger{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/change-password")
	if err != nil {
		t.Fatalf("requesting change password page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /change-password status = %d, want 200", resp.StatusCode)
	}
}

func TestChangePasswordPostSuccess(t *testing.T) {
	hash, _ := auth.HashPassword("old-password")

	users := &stubPasswordChanger{
		user: db.User{ID: 1, PasswordHash: hash},
	}

	sm := scs.New()
	srv := changePasswordTestServer(sm, users)
	defer srv.Close()

	form := url.Values{
		"current_password": {"old-password"},
		"new_password":     {"new-password"},
	}
	resp := authenticatedPost(t, srv, "/change-password", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with success)", resp.StatusCode)
	}
	if !users.updateCalled {
		t.Fatal("UpdateUserPassword was not called")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(users.updateParams.PasswordHash), []byte("new-password")); err != nil {
		t.Errorf("new password hash does not match: %v", err)
	}
}

func TestChangePasswordPostWrongCurrentPassword(t *testing.T) {
	hash, _ := auth.HashPassword("real-password")

	users := &stubPasswordChanger{
		user: db.User{ID: 1, PasswordHash: hash},
	}

	sm := scs.New()
	srv := changePasswordTestServer(sm, users)
	defer srv.Close()

	form := url.Values{
		"current_password": {"wrong-password"},
		"new_password":     {"new-password"},
	}
	resp := authenticatedPost(t, srv, "/change-password", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 (re-render with error)", resp.StatusCode)
	}
	if users.updateCalled {
		t.Error("UpdateUserPassword should not have been called")
	}
}
