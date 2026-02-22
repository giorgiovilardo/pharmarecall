package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

func TestLogoutDestroysSessionAndRedirects(t *testing.T) {
	sm := scs.New()

	mux := http.NewServeMux()
	// Setup route to create a session
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "personnel")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("POST /logout", web.HandleLogout(sm))
	// A protected-like route that checks if session has userID
	mux.HandleFunc("GET /check-session", func(w http.ResponseWriter, r *http.Request) {
		if sm.GetInt64(r.Context(), "userID") != 0 {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	srv := httptest.NewServer(sm.LoadAndSave(mux))
	defer srv.Close()

	client := noFollowClient()

	// Create a session
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()
	cookies := setupResp.Cookies()

	// Logout
	req, err := http.NewRequest(http.MethodPost, srv.URL+"/logout", nil)
	if err != nil {
		t.Fatalf("creating logout request: %v", err)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	logoutResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("posting logout: %v", err)
	}
	defer logoutResp.Body.Close()

	if logoutResp.StatusCode != http.StatusSeeOther {
		t.Errorf("POST /logout status = %d, want 303", logoutResp.StatusCode)
	}
	if loc := logoutResp.Header.Get("Location"); loc != "/login" {
		t.Errorf("POST /logout redirect = %q, want /login", loc)
	}

	// Verify session is destroyed — use the cookie from logout response
	logoutCookies := logoutResp.Cookies()
	checkReq, err := http.NewRequest(http.MethodGet, srv.URL+"/check-session", nil)
	if err != nil {
		t.Fatalf("creating check request: %v", err)
	}
	for _, c := range logoutCookies {
		checkReq.AddCookie(c)
	}

	checkResp, err := client.Do(checkReq)
	if err != nil {
		t.Fatalf("checking session: %v", err)
	}
	defer checkResp.Body.Close()

	if checkResp.StatusCode != http.StatusUnauthorized {
		t.Errorf("session check after logout status = %d, want 401 (session should be destroyed)", checkResp.StatusCode)
	}
}
