package notification

import "context"

// NotificationCreator creates a notification (ON CONFLICT DO NOTHING).
type NotificationCreator interface {
	Create(ctx context.Context, pharmacyID, prescriptionID int64, transitionType string) error
}

// NotificationLister lists notifications for a pharmacy.
type NotificationLister interface {
	ListByPharmacy(ctx context.Context, pharmacyID int64) ([]Notification, error)
}

// NotificationReader marks a single notification as read.
type NotificationReader interface {
	MarkRead(ctx context.Context, id, pharmacyID int64) error
}

// AllNotificationsReader marks all notifications as read for a pharmacy.
type AllNotificationsReader interface {
	MarkAllRead(ctx context.Context, pharmacyID int64) error
}

// UnreadCounter counts unread notifications for a pharmacy.
type UnreadCounter interface {
	CountUnread(ctx context.Context, pharmacyID int64) (int64, error)
}

// Repository composes all ports — used only by NewService for convenient wiring.
type Repository interface {
	NotificationCreator
	NotificationLister
	NotificationReader
	AllNotificationsReader
	UnreadCounter
}
