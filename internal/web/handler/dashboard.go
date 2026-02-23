package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/giorgiovilardo/pharmarecall/internal/order"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// OrderEnsurer creates pending orders for prescriptions in the lookahead window.
type OrderEnsurer interface {
	EnsureOrders(ctx context.Context, pharmacyID int64, now time.Time, lookaheadDays int) error
}

// DashboardLister lists dashboard entries for a pharmacy.
type DashboardLister interface {
	ListDashboard(ctx context.Context, pharmacyID int64) ([]order.DashboardEntry, error)
}

// OrderStatusAdvancer advances an order to the next status.
type OrderStatusAdvancer interface {
	AdvanceStatus(ctx context.Context, orderID int64, now time.Time) error
}

// DashboardFilters holds parsed filter parameters.
type DashboardFilters struct {
	PrescriptionStatus string
	OrderStatus        string
	DateFrom           string
	DateTo             string
}

// HandleDashboard renders the order dashboard for pharmacy staff.
func HandleDashboard(ensurer OrderEnsurer, lister DashboardLister, notifier ApproachingNotifier, lookaheadDays int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID := web.PharmacyID(r.Context())
		now := time.Now()

		// Ensure orders are created for prescriptions in the window.
		if err := ensurer.EnsureOrders(r.Context(), pharmacyID, now, lookaheadDays); err != nil {
			slog.Error("ensuring orders", "error", err)
		}

		entries, err := lister.ListDashboard(r.Context(), pharmacyID)
		if err != nil {
			slog.Error("listing dashboard", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		// Generate notifications for approaching prescriptions.
		var approachingIDs []int64
		for _, e := range entries {
			if e.PrescriptionStatus(now) == "approaching" {
				approachingIDs = append(approachingIDs, e.PrescriptionID)
			}
		}
		if len(approachingIDs) > 0 {
			if err := notifier.GenerateApproaching(r.Context(), pharmacyID, approachingIDs); err != nil {
				slog.Error("generating notifications", "error", err)
			}
		}

		filters := DashboardFilters{
			PrescriptionStatus: r.URL.Query().Get("rx_status"),
			OrderStatus:        r.URL.Query().Get("order_status"),
			DateFrom:           r.URL.Query().Get("date_from"),
			DateTo:             r.URL.Query().Get("date_to"),
		}

		filtered := applyDashboardFilters(entries, filters, now)

		web.OrderDashboardPage(filtered, now, filters.PrescriptionStatus, filters.OrderStatus, filters.DateFrom, filters.DateTo).Render(r.Context(), w)
	}
}

// HandleAdvanceOrderStatus advances an order to the next status in its lifecycle.
func HandleAdvanceOrderStatus(advancer OrderStatusAdvancer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := advancer.AdvanceStatus(r.Context(), orderID, time.Now().Truncate(24*time.Hour)); err != nil {
			if errors.Is(err, order.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			if errors.Is(err, order.ErrInvalidTransition) {
				http.Error(w, "Transizione di stato non valida.", http.StatusBadRequest)
				return
			}
			slog.Error("advancing order status", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		// Preserve filters when redirecting back.
		redirectURL := "/dashboard"
		if r.URL.RawQuery != "" {
			redirectURL += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

// HandlePrintDashboard renders a print-friendly version of the order dashboard.
func HandlePrintDashboard(lister DashboardLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID := web.PharmacyID(r.Context())
		now := time.Now()

		entries, err := lister.ListDashboard(r.Context(), pharmacyID)
		if err != nil {
			slog.Error("listing dashboard for print", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		filters := DashboardFilters{
			PrescriptionStatus: r.URL.Query().Get("rx_status"),
			OrderStatus:        r.URL.Query().Get("order_status"),
			DateFrom:           r.URL.Query().Get("date_from"),
			DateTo:             r.URL.Query().Get("date_to"),
		}

		filtered := applyDashboardFilters(entries, filters, now)

		pharmacyName := web.PharmacyName(r.Context())
		web.PrintDashboardPage(filtered, now, pharmacyName).Render(r.Context(), w)
	}
}

// HandlePrintLabel renders a print-friendly label for a single order.
func HandlePrintLabel(lister DashboardLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		pharmacyID := web.PharmacyID(r.Context())

		entries, err := lister.ListDashboard(r.Context(), pharmacyID)
		if err != nil {
			slog.Error("listing dashboard for label", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		var found *order.DashboardEntry
		for _, e := range entries {
			if e.OrderID == orderID {
				found = &e
				break
			}
		}
		if found == nil {
			http.NotFound(w, r)
			return
		}

		web.PrintLabelsPage([]order.DashboardEntry{*found}).Render(r.Context(), w)
	}
}

// HandlePrintBatchLabels renders print-friendly labels for all filtered orders.
func HandlePrintBatchLabels(lister DashboardLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID := web.PharmacyID(r.Context())
		now := time.Now()

		entries, err := lister.ListDashboard(r.Context(), pharmacyID)
		if err != nil {
			slog.Error("listing dashboard for batch labels", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		filters := DashboardFilters{
			PrescriptionStatus: r.URL.Query().Get("rx_status"),
			OrderStatus:        r.URL.Query().Get("order_status"),
			DateFrom:           r.URL.Query().Get("date_from"),
			DateTo:             r.URL.Query().Get("date_to"),
		}

		filtered := applyDashboardFilters(entries, filters, now)

		web.PrintLabelsPage(filtered).Render(r.Context(), w)
	}
}

func applyDashboardFilters(entries []order.DashboardEntry, filters DashboardFilters, now time.Time) []order.DashboardEntry {
	var result []order.DashboardEntry

	dateFrom, _ := time.Parse("2006-01-02", filters.DateFrom)
	dateTo, _ := time.Parse("2006-01-02", filters.DateTo)

	for _, e := range entries {
		if filters.PrescriptionStatus != "" && filters.PrescriptionStatus != "all" {
			if e.PrescriptionStatus(now) != filters.PrescriptionStatus {
				continue
			}
		}

		switch filters.OrderStatus {
		case "all":
			// Show everything, no filtering.
		case "":
			// Default: show only active orders (pending + prepared).
			if e.OrderStatus == order.StatusFulfilled {
				continue
			}
		default:
			// Explicit single-status filter.
			if e.OrderStatus != filters.OrderStatus {
				continue
			}
		}

		if !dateFrom.IsZero() {
			depletion := e.EstimatedDepletionDate.Truncate(24 * time.Hour)
			if depletion.Before(dateFrom) {
				continue
			}
		}

		if !dateTo.IsZero() {
			depletion := e.EstimatedDepletionDate.Truncate(24 * time.Hour)
			if depletion.After(dateTo) {
				continue
			}
		}

		result = append(result, e)
	}

	return result
}
