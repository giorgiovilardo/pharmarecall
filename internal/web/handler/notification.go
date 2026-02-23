package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/giorgiovilardo/pharmarecall/internal/notification"
	"github.com/giorgiovilardo/pharmarecall/internal/web"
)

// ApproachingNotifier generates notifications for prescriptions entering approaching status.
type ApproachingNotifier interface {
	GenerateApproaching(ctx context.Context, pharmacyID int64, prescriptionIDs []int64) error
}

// NotificationLister lists notifications for a pharmacy.
type NotificationLister interface {
	List(ctx context.Context, pharmacyID int64) ([]notification.Notification, error)
}

// NotificationMarkReader marks a single notification as read.
type NotificationMarkReader interface {
	MarkRead(ctx context.Context, id, pharmacyID int64) error
}

// NotificationMarkAllReader marks all notifications as read.
type NotificationMarkAllReader interface {
	MarkAllRead(ctx context.Context, pharmacyID int64) error
}

// HandleNotificationList renders the notification list page.
func HandleNotificationList(lister NotificationLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID := web.PharmacyID(r.Context())

		notifs, err := lister.List(r.Context(), pharmacyID)
		if err != nil {
			slog.Error("listing notifications", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		web.NotificationListPage(notifs).Render(r.Context(), w)
	}
}

// HandleMarkNotificationRead marks a single notification as read.
func HandleMarkNotificationRead(reader NotificationMarkReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		pharmacyID := web.PharmacyID(r.Context())

		if err := reader.MarkRead(r.Context(), id, pharmacyID); err != nil {
			slog.Error("marking notification read", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/notifications", http.StatusSeeOther)
	}
}

// HandleMarkAllNotificationsRead marks all notifications as read.
func HandleMarkAllNotificationsRead(reader NotificationMarkAllReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pharmacyID := web.PharmacyID(r.Context())

		if err := reader.MarkAllRead(r.Context(), pharmacyID); err != nil {
			slog.Error("marking all notifications read", "error", err)
			http.Error(w, "Errore interno.", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/notifications", http.StatusSeeOther)
	}
}

