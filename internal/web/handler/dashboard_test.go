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

type stubApproachingNotifier struct {
	called          bool
	pharmacyID      int64
	prescriptionIDs []int64
	err             error
}

func (s *stubApproachingNotifier) GenerateApproaching(_ context.Context, pharmacyID int64, prescriptionIDs []int64) error {
	s.called = true
	s.pharmacyID = pharmacyID
	s.prescriptionIDs = prescriptionIDs
	return s.err
}

// --- Dashboard test server ---

type dashTestDeps struct {
	sm       *scs.SessionManager
	ensurer  handler.OrderEnsurer
	lister   handler.DashboardLister
	notifier handler.ApproachingNotifier
	advancer handler.OrderStatusAdvancer
}

func dashTestServer(d dashTestDeps) *httptest.Server {
	mux := http.NewServeMux()
	if d.ensurer != nil && d.lister != nil {
		notifier := d.notifier
		if notifier == nil {
			notifier = &stubApproachingNotifier{}
		}
		mux.Handle("GET /dashboard", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandleDashboard(d.ensurer, d.lister, notifier, 7))))
		mux.Handle("GET /dashboard/print", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandlePrintDashboard(d.lister))))
		mux.Handle("GET /dashboard/labels", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandlePrintBatchLabels(d.lister))))
		mux.Handle("GET /orders/{id}/label", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandlePrintLabel(d.lister))))
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

func TestDashboardDefaultsToActiveOrders(t *testing.T) {
	ensurer := &stubOrderEnsurer{}
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending, FirstName: "A", LastName: "A"},
		{OrderID: 2, MedicationName: "Aspirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusFulfilled, FirstName: "B", LastName: "B"},
		{OrderID: 3, MedicationName: "Moment", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPrepared, FirstName: "C", LastName: "C"},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: ensurer, lister: lister})
	defer srv.Close()

	// No filter → should show only pending + prepared, not fulfilled
	resp := authenticatedGet(t, srv, "/dashboard")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Tachipirina") {
		t.Error("body should contain Tachipirina (pending)")
	}
	if !strings.Contains(bodyStr, "Moment") {
		t.Error("body should contain Moment (prepared)")
	}
	if strings.Contains(bodyStr, "Aspirina") {
		t.Error("body should not contain Aspirina (fulfilled, hidden by default)")
	}
}

func TestDashboardShowsFulfilledWhenExplicitlyFiltered(t *testing.T) {
	ensurer := &stubOrderEnsurer{}
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending, FirstName: "A", LastName: "A"},
		{OrderID: 2, MedicationName: "Aspirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusFulfilled, FirstName: "B", LastName: "B"},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: ensurer, lister: lister})
	defer srv.Close()

	// Explicit "all" → should show everything including fulfilled
	resp := authenticatedGet(t, srv, "/dashboard?order_status=all")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Tachipirina") {
		t.Error("body should contain Tachipirina (pending)")
	}
	if !strings.Contains(bodyStr, "Aspirina") {
		t.Error("body should contain Aspirina (fulfilled, explicitly showing all)")
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

// --- Print dashboard tests (9.1, 9.2, 9.3) ---

func TestPrintDashboardTriggersWindowPrint(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", EstimatedDepletionDate: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending, FirstName: "A", LastName: "A"},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard/print")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "window.print()") {
		t.Error("print page should contain window.print() trigger")
	}
}

func TestPrintLabelsTriggersWindowPrint(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", FirstName: "A", LastName: "A", Fulfillment: "pickup", EstimatedDepletionDate: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard/labels")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "window.print()") {
		t.Error("labels page should contain window.print() trigger")
	}
}

func TestPrintDashboardRendersEntriesWithoutNav(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{
			OrderID:                1,
			PrescriptionID:         10,
			EstimatedDepletionDate: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
			OrderStatus:            order.StatusPending,
			MedicationName:         "Tachipirina",
			UnitsPerBox:            30,
			PatientID:              100,
			FirstName:              "Mario",
			LastName:               "Rossi",
			Fulfillment:            "pickup",
		},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard/print")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Should contain order data
	for _, want := range []string{"Tachipirina", "Mario", "Rossi", "30"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}

	// Should NOT contain navigation elements (no Layout wrapping)
	if strings.Contains(bodyStr, "data-topnav") {
		t.Error("print view should not contain navigation")
	}
}

func TestPrintDashboardRespectsFilters(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending, FirstName: "A", LastName: "A"},
		{OrderID: 2, MedicationName: "Aspirina", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPrepared, FirstName: "B", LastName: "B"},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard/print?order_status=prepared")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Aspirina") {
		t.Error("print view should contain Aspirina (prepared)")
	}
	if strings.Contains(bodyStr, "Tachipirina") {
		t.Error("print view should not contain Tachipirina (pending, filtered out)")
	}
}

// --- Label print tests (9.4, 9.5, 9.6) ---

func TestPrintSingleLabelShippingIncludesAddress(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{
			OrderID:                5,
			MedicationName:         "Tachipirina",
			FirstName:              "Mario",
			LastName:               "Rossi",
			Fulfillment:            "shipping",
			DeliveryAddress:        "Via Roma 1, Milano",
			Phone:                  "3331234567",
			EstimatedDepletionDate: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
			OrderStatus:            order.StatusPending,
		},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/orders/5/label")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Mario", "Rossi", "Tachipirina", "Via Roma 1, Milano", "Spedizione"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}

	if strings.Contains(bodyStr, "data-topnav") {
		t.Error("label view should not contain navigation")
	}
}

func TestPrintSingleLabelPickupIncludesContact(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{
			OrderID:                3,
			MedicationName:         "Aspirina",
			FirstName:              "Luca",
			LastName:               "Bianchi",
			Fulfillment:            "pickup",
			Phone:                  "3339876543",
			Email:                  "luca@example.com",
			EstimatedDepletionDate: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC),
			OrderStatus:            order.StatusPrepared,
		},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/orders/3/label")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Luca", "Bianchi", "Aspirina", "3339876543", "Ritiro"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestPrintSingleLabelNotFoundReturns404(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", FirstName: "A", LastName: "A", EstimatedDepletionDate: time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/orders/999/label")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPrintBatchLabelsRendersAllFilteredEntries(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", FirstName: "Mario", LastName: "Rossi", Fulfillment: "pickup", Phone: "333111", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending},
		{OrderID: 2, MedicationName: "Aspirina", FirstName: "Luca", LastName: "Bianchi", Fulfillment: "shipping", DeliveryAddress: "Via Dante 5", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPrepared},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard/labels")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Tachipirina", "Mario", "Aspirina", "Luca", "Via Dante 5", "333111"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}

	if strings.Contains(bodyStr, "data-topnav") {
		t.Error("label view should not contain navigation")
	}
}

func TestPrintBatchLabelsRespectsFilters(t *testing.T) {
	lister := &stubDashboardLister{result: []order.DashboardEntry{
		{OrderID: 1, MedicationName: "Tachipirina", FirstName: "A", LastName: "A", Fulfillment: "pickup", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPending},
		{OrderID: 2, MedicationName: "Aspirina", FirstName: "B", LastName: "B", Fulfillment: "shipping", EstimatedDepletionDate: time.Date(2026, 2, 25, 0, 0, 0, 0, time.UTC), OrderStatus: order.StatusPrepared},
	}}

	sm := scs.New()
	srv := dashTestServer(dashTestDeps{sm: sm, ensurer: &stubOrderEnsurer{}, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/dashboard/labels?order_status=prepared")
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "Aspirina") {
		t.Error("batch labels should contain Aspirina (prepared)")
	}
	if strings.Contains(bodyStr, "Tachipirina") {
		t.Error("batch labels should not contain Tachipirina (pending, filtered out)")
	}
}
