package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

func TestRequireAuthRedirectsUnauthenticated(t *testing.T) {
	sm := scs.New()

	protected := web.RequireAuth(sm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	srv := httptest.NewServer(sm.LoadAndSave(protected))
	defer srv.Close()

	resp, err := noFollowClient().Get(srv.URL + "/protected")
	if err != nil {
		t.Fatalf("requesting protected page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Errorf("redirect = %q, want /login", loc)
	}
}

func TestRequireAuthAllowsAuthenticatedAndSetsContext(t *testing.T) {
	sm := scs.New()

	var gotUserID int64
	var gotRole string

	protected := web.RequireAuth(sm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = web.UserID(r.Context())
		gotRole = web.Role(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	mux := http.NewServeMux()
	mux.Handle("GET /protected", protected)
	// Setup route to create a session (simulates login)
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(42))
		sm.Put(r.Context(), "role", "personnel")
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(mux))
	defer srv.Close()

	// First, set up a session by hitting the setup route
	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	// Extract session cookie
	cookies := setupResp.Cookies()

	// Now hit the protected route with the session cookie
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/protected", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting protected page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if gotUserID != 42 {
		t.Errorf("userID = %d, want 42", gotUserID)
	}
	if gotRole != "personnel" {
		t.Errorf("role = %q, want personnel", gotRole)
	}
}
