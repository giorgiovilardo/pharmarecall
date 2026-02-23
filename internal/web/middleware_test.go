package web_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// --- Notification count mock ---

type stubUnreadCounter struct {
	result int64
	err    error
}

func (s *stubUnreadCounter) CountUnread(_ context.Context, _ int64) (int64, error) {
	return s.result, s.err
}

func TestRequireAuthRedirectsUnauthenticated(t *testing.T) {
	sm := scs.New()

	protected := web.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// LoadUser + RequireAuth chain
	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(protected)))
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

func TestLoadUserSetsContextForAuthenticatedUser(t *testing.T) {
	sm := scs.New()

	var gotUserID int64
	var gotRole string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = web.UserID(r.Context())
		gotRole = web.Role(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /check", handler)
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(42))
		sm.Put(r.Context(), "role", "personnel")
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()
	cookies := setupResp.Cookies()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting page: %v", err)
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

func TestRequireAdminAllowsAdminRole(t *testing.T) {
	sm := scs.New()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /admin-only", web.RequireAdmin(handler))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/admin-only", nil)
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting admin page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRequireAdminDeniesNonAdminRoles(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{"owner denied", "owner"},
		{"personnel denied", "personnel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := scs.New()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			mux := http.NewServeMux()
			mux.Handle("GET /admin-only", web.RequireAdmin(handler))
			mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
				sm.Put(r.Context(), "userID", int64(1))
				sm.Put(r.Context(), "role", tt.role)
				w.WriteHeader(http.StatusOK)
			})

			srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
			defer srv.Close()

			client := noFollowClient()
			setupResp, err := client.Get(srv.URL + "/setup-session")
			if err != nil {
				t.Fatalf("setting up session: %v", err)
			}
			setupResp.Body.Close()

			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/admin-only", nil)
			for _, c := range setupResp.Cookies() {
				req.AddCookie(c)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("requesting admin page: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("status = %d, want 403", resp.StatusCode)
			}
		})
	}
}

func TestRequireOwnerAllowsOwnerRole(t *testing.T) {
	sm := scs.New()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /owner-only", web.RequireOwner(handler))
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "owner")
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/owner-only", nil)
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting owner page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRequireOwnerDeniesNonOwnerRoles(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{"admin denied", "admin"},
		{"personnel denied", "personnel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := scs.New()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			mux := http.NewServeMux()
			mux.Handle("GET /owner-only", web.RequireOwner(handler))
			mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
				sm.Put(r.Context(), "userID", int64(1))
				sm.Put(r.Context(), "role", tt.role)
				w.WriteHeader(http.StatusOK)
			})

			srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
			defer srv.Close()

			client := noFollowClient()
			setupResp, err := client.Get(srv.URL + "/setup-session")
			if err != nil {
				t.Fatalf("setting up session: %v", err)
			}
			setupResp.Body.Close()

			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/owner-only", nil)
			for _, c := range setupResp.Cookies() {
				req.AddCookie(c)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("requesting owner page: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("status = %d, want 403", resp.StatusCode)
			}
		})
	}
}

func TestLoadUserSetsPharmacyIDInContext(t *testing.T) {
	sm := scs.New()

	var gotPharmacyID int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPharmacyID = web.PharmacyID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /check", handler)
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(42))
		sm.Put(r.Context(), "role", "owner")
		sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if gotPharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", gotPharmacyID)
	}
}

func TestLoadUserSetsUserNameAndPharmacyName(t *testing.T) {
	sm := scs.New()

	var gotUserName, gotPharmacyName string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserName = web.UserName(r.Context())
		gotPharmacyName = web.PharmacyName(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /check", handler)
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(42))
		sm.Put(r.Context(), "role", "owner")
		sm.Put(r.Context(), "pharmacyID", int64(7))
		sm.Put(r.Context(), "userName", "Mario Rossi")
		sm.Put(r.Context(), "pharmacyName", "Farmacia Centrale")
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(mux)))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting page: %v", err)
	}
	defer resp.Body.Close()

	if gotUserName != "Mario Rossi" {
		t.Errorf("userName = %q, want %q", gotUserName, "Mario Rossi")
	}
	if gotPharmacyName != "Farmacia Centrale" {
		t.Errorf("pharmacyName = %q, want %q", gotPharmacyName, "Farmacia Centrale")
	}
}

func TestLoadUserPassesThroughForUnauthenticated(t *testing.T) {
	sm := scs.New()

	var gotUserID int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = web.UserID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(handler)))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("requesting page: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if gotUserID != 0 {
		t.Errorf("userID = %d, want 0 (unauthenticated)", gotUserID)
	}
}

// --- LoadNotificationCount tests ---

func TestLoadNotificationCountSetsCountForPharmacyStaff(t *testing.T) {
	sm := scs.New()
	counter := &stubUnreadCounter{result: 5}

	var gotCount int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCount = web.UnreadNotificationCount(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /check", handler)
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "personnel")
		sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(web.LoadNotificationCount(counter)(mux))))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting page: %v", err)
	}
	defer resp.Body.Close()

	if gotCount != 5 {
		t.Errorf("unreadCount = %d, want 5", gotCount)
	}
}

func TestLoadNotificationCountSkipsForAdmin(t *testing.T) {
	sm := scs.New()
	counter := &stubUnreadCounter{result: 10}

	var gotCount int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCount = web.UnreadNotificationCount(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /check", handler)
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "admin")
		// admin has no pharmacyID
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(web.LoadNotificationCount(counter)(mux))))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting page: %v", err)
	}
	defer resp.Body.Close()

	if gotCount != 0 {
		t.Errorf("unreadCount = %d, want 0 (admin has no pharmacy)", gotCount)
	}
}

func TestLoadNotificationCountZeroNotSetInContext(t *testing.T) {
	sm := scs.New()
	counter := &stubUnreadCounter{result: 0}

	var gotCount int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCount = web.UnreadNotificationCount(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	mux := http.NewServeMux()
	mux.Handle("GET /check", handler)
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		sm.Put(r.Context(), "userID", int64(1))
		sm.Put(r.Context(), "role", "personnel")
		sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(web.LoadNotificationCount(counter)(mux))))
	defer srv.Close()

	client := noFollowClient()
	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/check", nil)
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting page: %v", err)
	}
	defer resp.Body.Close()

	if gotCount != 0 {
		t.Errorf("unreadCount = %d, want 0", gotCount)
	}
}

func TestLoadNotificationCountSkipsForUnauthenticated(t *testing.T) {
	sm := scs.New()
	counter := &stubUnreadCounter{result: 3}

	var gotCount int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCount = web.UnreadNotificationCount(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	srv := httptest.NewServer(sm.LoadAndSave(web.LoadUser(sm)(web.LoadNotificationCount(counter)(handler))))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("requesting page: %v", err)
	}
	defer resp.Body.Close()

	if gotCount != 0 {
		t.Errorf("unreadCount = %d, want 0 (unauthenticated)", gotCount)
	}
}
