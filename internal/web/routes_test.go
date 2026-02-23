package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// newTestStack builds a full handler stack (router + sessions + CORS) for
// integration-level route tests.
func noopHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newTestStack() http.Handler {
	sm := scs.New()
	stub := &stubUserGetter{}
	mux := web.NewRouter(
		web.HandleLoginPage(),
		web.HandleLoginPost(sm, stub),
		web.HandleLogout(sm),
		web.HandleChangePasswordPage(),
		web.HandleChangePasswordPost(sm, &stubPasswordChanger{}),
		web.AdminHandlers{
			Dashboard:       noopHandler,
			NewPharmacy:     noopHandler,
			CreatePharmacy:  noopHandler,
			PharmacyDetail:  noopHandler,
			UpdatePharmacy:  noopHandler,
			AddPersonnel:    noopHandler,
			CreatePersonnel: noopHandler,
		},
	)
	cop := http.NewCrossOriginProtection()
	return cop.Handler(sm.LoadAndSave(web.LoadUser(sm)(mux)))
}

func TestCrossOriginPostRejected(t *testing.T) {
	srv := httptest.NewServer(newTestStack())
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("performing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-origin POST status = %d, want 403", resp.StatusCode)
	}
}

func TestHealthCheck(t *testing.T) {
	srv := httptest.NewServer(newTestStack())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("requesting health check: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET / status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("GET / content-type = %q, want text/html; charset=utf-8", ct)
	}
}

func TestStaticFileServing(t *testing.T) {
	srv := httptest.NewServer(newTestStack())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/custom.css")
	if err != nil {
		t.Fatalf("requesting static file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /static/custom.css status = %d, want 200", resp.StatusCode)
	}
}
