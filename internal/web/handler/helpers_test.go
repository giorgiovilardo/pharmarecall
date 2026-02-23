package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// noFollowClient returns an HTTP client that does not follow redirects.
func noFollowClient() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

// authenticatedPost creates a session then posts to the given URL with form data.
func authenticatedPost(t *testing.T, srv *httptest.Server, path string, form url.Values) *http.Response {
	t.Helper()
	client := noFollowClient()

	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()
	cookies := setupResp.Cookies()

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

// authenticatedGet creates a session then GETs the given URL.
func authenticatedGet(t *testing.T, srv *httptest.Server, path string) *http.Response {
	t.Helper()
	client := noFollowClient()

	setupResp, err := client.Get(srv.URL + "/setup-session")
	if err != nil {
		t.Fatalf("setting up session: %v", err)
	}
	setupResp.Body.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	for _, c := range setupResp.Cookies() {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("requesting %s: %v", path, err)
	}
	return resp
}
