package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/giorgiovilardo/pharmarecall/internal/notification"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
	"github.com/giorgiovilardo/pharmarecall/internal/web/handler"
)

// --- Notification stubs ---

type stubNotificationLister struct {
	result []notification.Notification
	err    error
}

func (s *stubNotificationLister) List(_ context.Context, _ int64) ([]notification.Notification, error) {
	return s.result, s.err
}

type stubNotificationMarkReader struct {
	called     bool
	id         int64
	pharmacyID int64
	err        error
}

func (s *stubNotificationMarkReader) MarkRead(_ context.Context, id, pharmacyID int64) error {
	s.called = true
	s.id = id
	s.pharmacyID = pharmacyID
	return s.err
}

type stubNotificationMarkAllReader struct {
	called     bool
	pharmacyID int64
	err        error
}

func (s *stubNotificationMarkAllReader) MarkAllRead(_ context.Context, pharmacyID int64) error {
	s.called = true
	s.pharmacyID = pharmacyID
	return s.err
}

// --- Notification test server ---

type notifTestDeps struct {
	sm       *scs.SessionManager
	lister   handler.NotificationLister
	markRead handler.NotificationMarkReader
	markAll  handler.NotificationMarkAllReader
}

func notifTestServer(d notifTestDeps) *httptest.Server {
	mux := http.NewServeMux()
	if d.lister != nil {
		mux.Handle("GET /notifications", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandleNotificationList(d.lister))))
	}
	if d.markRead != nil {
		mux.Handle("POST /notifications/{id}/read", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandleMarkNotificationRead(d.markRead))))
	}
	if d.markAll != nil {
		mux.Handle("POST /notifications/read-all", web.RequirePharmacyStaff(http.HandlerFunc(handler.HandleMarkAllNotificationsRead(d.markAll))))
	}
	mux.HandleFunc("GET /setup-session", func(w http.ResponseWriter, r *http.Request) {
		d.sm.Put(r.Context(), "userID", int64(1))
		d.sm.Put(r.Context(), "role", "personnel")
		d.sm.Put(r.Context(), "pharmacyID", int64(7))
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(d.sm.LoadAndSave(web.LoadUser(d.sm)(mux)))
}

// --- Notification list tests ---

func TestNotificationListRendersEntries(t *testing.T) {
	lister := &stubNotificationLister{result: []notification.Notification{
		{
			ID:               1,
			PharmacyID:       7,
			PrescriptionID:   10,
			TransitionType:   notification.TransitionApproaching,
			Read:             false,
			CreatedAt:        time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			MedicationName:   "Tachipirina",
			UnitsPerBox:      30,
			DailyConsumption: 1,
			BoxStartDate:     time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			PatientID:        100,
			FirstName:        "Mario",
			LastName:         "Rossi",
		},
	}}

	sm := scs.New()
	srv := notifTestServer(notifTestDeps{sm: sm, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/notifications")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	for _, want := range []string{"Tachipirina", "Mario", "Rossi", "Non letta"} {
		if !strings.Contains(bodyStr, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestNotificationListEmptyShowsMessage(t *testing.T) {
	lister := &stubNotificationLister{result: nil}

	sm := scs.New()
	srv := notifTestServer(notifTestDeps{sm: sm, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/notifications")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Nessuna notifica") {
		t.Error("body missing empty message")
	}
}

func TestNotificationListErrorReturns500(t *testing.T) {
	lister := &stubNotificationLister{err: errors.New("db down")}

	sm := scs.New()
	srv := notifTestServer(notifTestDeps{sm: sm, lister: lister})
	defer srv.Close()

	resp := authenticatedGet(t, srv, "/notifications")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}

// --- Mark as read tests ---

func TestMarkNotificationReadRedirects(t *testing.T) {
	reader := &stubNotificationMarkReader{}

	sm := scs.New()
	srv := notifTestServer(notifTestDeps{sm: sm, markRead: reader})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/notifications/42/read", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/notifications" {
		t.Errorf("redirect = %q, want /notifications", loc)
	}
	if !reader.called {
		t.Error("MarkRead was not called")
	}
	if reader.id != 42 {
		t.Errorf("id = %d, want 42", reader.id)
	}
	if reader.pharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", reader.pharmacyID)
	}
}

func TestMarkNotificationReadInvalidIDReturns404(t *testing.T) {
	reader := &stubNotificationMarkReader{}

	sm := scs.New()
	srv := notifTestServer(notifTestDeps{sm: sm, markRead: reader})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/notifications/abc/read", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

// --- Mark all as read tests ---

func TestMarkAllNotificationsReadRedirects(t *testing.T) {
	reader := &stubNotificationMarkAllReader{}

	sm := scs.New()
	srv := notifTestServer(notifTestDeps{sm: sm, markAll: reader})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/notifications/read-all", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/notifications" {
		t.Errorf("redirect = %q, want /notifications", loc)
	}
	if !reader.called {
		t.Error("MarkAllRead was not called")
	}
	if reader.pharmacyID != 7 {
		t.Errorf("pharmacyID = %d, want 7", reader.pharmacyID)
	}
}

func TestMarkAllNotificationsReadErrorReturns500(t *testing.T) {
	reader := &stubNotificationMarkAllReader{err: errors.New("db down")}

	sm := scs.New()
	srv := notifTestServer(notifTestDeps{sm: sm, markAll: reader})
	defer srv.Close()

	resp := authenticatedPost(t, srv, "/notifications/read-all", url.Values{})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", resp.StatusCode)
	}
}
