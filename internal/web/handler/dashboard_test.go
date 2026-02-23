package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/order"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

// --- Dashboard stubs ---

type stubOrderEnsurer struct {
	called     bool
	pharmacyID int64
	err        error
}

func (s *stubOrderEnsurer) EnsureOrders(_ context.Context, pharmacyID int64, _ time.Time, _ int) error {
	s.called = true
	s.pharmacyID = pharmacyID
	return s.err
}

type stubDashboardLister struct {
	result []order.DashboardEntry
	err    error
}

func (s *stubDashboardLister) ListDashboard(_ context.Context, _ int64) ([]order.DashboardEntry, error) {
	return s.result, s.err
}

type stubOrderAdvancer struct {
	called  bool
	orderID int64
	now     time.Time
	err     error
}

func (s *stubOrderAdvancer) AdvanceStatus(_ context.Context, orderID int64, now time.Time) error {
	s.called = true
	s.orderID = orderID
	s.now = now
	return s.err
}

// --- Dashboard test server ---

type dashTestDeps struct {
	sm       *scs.SessionManager
	ensurer  handler.OrderEnsurer
	lister   handler.DashboardLister
	advancer handler.OrderStatusAdvancer
}

func dashTestServer(d dashTestDeps) *httptest.Server {
	mux := http.NewServeMux()
	if d.ensurer != nil && d.lister != nil {
		mux.Handle("GET /dashboard", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandleDashboard(d.ensurer, d.lister, 7))))
	}
	if d.advancer != nil {
		mux.Handle("POST /orders/{id}/advance", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandleAdvanceOrderStatus(d.advancer))))
	}
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		d.sm.Put(r.Context(), "userID", int64(1))
		d.sm.Put(r.Context(), "role", "personnel")
		d.sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(d.sm.LoadAndSave(web.LoadUser(d.sm)(mux)))
}

// --- Dashboard GET tests (7.4, 7.5) ---

func TestDashboardRendersEntries(t *testing.T) {
	ensurer := &stubOrderEnsurer{}
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{
			OrderID:                1,
			PrescriptionID:         10,
			EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC),
			OrderStatus:            order.StatusPending,
			MedicationName:         "Tachipirina",
			PatientID:              100,
			FirstName:              "Mario",
			LastName:               "Rossi",
			Fulfillment:            "pickup",
		},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: ensurer, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Tachipirina", "Mario", "Rossi", "pending"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}

	if !ensurer.called {
		t.Error("EnsureOrders was not called")
	}
	if ensurer.pharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", ensurer.pharmacyID)
	}
}

func TestDashboardEmptyShowsMessage(t *testing.T) {
	ensurer := &stubOrderEnsurer{}
	lister := &stubDashboardLister{result: nil}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: ensurer, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nessun ordine") {
		t.Error("body missing empty message")
	}
}

// --- Filter tests (7.6, 7.7, 7.8) ---

func TestDashboardFiltersByPrescriptionStatus(t *testing.T) {
	ensurer := &stubOrderEnsurer{}
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", EstimatedDepletionDate: time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending, FirstName: "Mario", LastName: "Rossi"},
		{OrderID: 2, MedicationName: "Aspirina", EstimatedDepletionDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending, FirstName: "Luca", LastName: "Bianchi"},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: ensurer, lister: lister})
	defer srv.Close()

	// Filter to "approaching" only — Aspirina (May 1) should be "ok", Tachipirina (Feb 20) should be "approaching" or "depleted"
	resp := authenticatedGet(t, srv, "/dashboard?rx_status=ok")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Aspirina has depletion May 1 → far in the future → "ok" → should appear
	if !strings.Contains(bodyStr, "Aspirina") {
		t.Error("body should contain Aspirina (status ok)")
	}
}

func TestDashboardFiltersByOrderStatus(t *testing.T) {
	ensurer := &stubOrderEnsurer{}
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending, FirstName: "A", LastName: "A"},
		{OrderID: 2, MedicationName: "Aspirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPrepared, FirstName: "B", LastName: "B"},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: ensurer, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard?order_status=prepared")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Aspirina") {
		t.Error("body should contain Aspirina (prepared)")
	}
	if strings.Contains(bodyStr, "Tachipirina") {
		t.Error("body should not contain Tachipirina (pending, filtered out)")
	}
}

// --- Advance order status tests (7.9) ---

func TestAdvanceOrderStatusRedirects(t *testing.T) {
	advancer := &stubOrderAdvancer{}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, advancer: advancer})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/orders/5/advance", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/dashboard" {
		t.Errorf("redirect = %q, want /dashboard", loc)
	}
	if !advancer.called {
		t.Error("AdvanceStatus was not called")
	}
	if advancer.orderID != 5 {
		t.Errorf("orderID = %d, want 5", advancer.orderID)
	}
}

func TestAdvanceOrderStatusNotFoundReturns404(t *testing.T) {
	advancer := &stubOrderAdvancer{err: order.ErrNotFound}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, advancer: advancer})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/orders/999/advance", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAdvanceOrderStatusInvalidTransitionReturns400(t *testing.T) {
	advancer := &stubOrderAdvancer{err: order.ErrInvalidTransition}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, advancer: advancer})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/orders/1/advance", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}
