package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/user"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

type stubPasswordChanger struct {
	called bool
	err    error
}

func (s *stubPasswordChanger) ChangePassword(_ context.Context, _ int64, _, _ string) error {
	s.called = true
	return s.err
}

func changePasswordTestServer(sm *scs.SessionManager, changer handler.PasswordChanger) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /change-password", handler.HandleChangePasswordPage())
	mux.HandleFunc("POST /change-password", handler.HandleChangePasswordPost(sm, changer))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(sm.LoadAndSave(mux))
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
	changer := &stubPasswordChanger{}

	sm := scs.New()
	srv := changePasswordTestServer(sm, changer)
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
	if !changer.called {
		t.Fatal("ChangePassword was not called")
	}
}

func TestChangePasswordPostWrongCurrentPassword(t *testing.T) {
	changer := &stubPasswordChanger{err: user.ErrInvalidCredentials}

	sm := scs.New()
	srv := changePasswordTestServer(sm, changer)
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
}
