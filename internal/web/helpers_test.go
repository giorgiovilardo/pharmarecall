package web_test

import "net/http"

// noFollowClient returns an HTTP client that does not follow redirects.
func noFollowClient() *http.Client {
	return &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}
