package notification

import (
	"context"
	"fmt"
)

// ServiceDeps holds individual port interfaces — used by tests to inject only what's needed.
type ServiceDeps struct {
	Creator   NotificationCreator
	Lister    NotificationLister
	Reader    NotificationReader
	AllReader AllNotificationsReader
	Counter   UnreadCounter
}

// Service contains notification domain business logic.
type Service struct {
	deps ServiceDeps
}

// NewService is the production constructor — takes a Repository (satisfies all ports).
func NewService(repo Repository) *Service {
	return &Service{deps: ServiceDeps{
		Creator:   repo,
		Lister:    repo,
		Reader:    repo,
		AllReader: repo,
		Counter:   repo,
	}}
}

// NewServiceWith is the test constructor — inject only what you need, rest stays nil.
func NewServiceWith(d ServiceDeps) *Service {
	return &Service{deps: d}
}

// GenerateApproaching creates notifications for prescriptions entering approaching status.
// Uses ON CONFLICT DO NOTHING at the DB level for idempotency.
func (s *Service) GenerateApproaching(ctx context.Context, pharmacyID int64, prescriptionIDs []int64) error {
	for _, rxID := range prescriptionIDs {
		if err := s.deps.Creator.Create(ctx, pharmacyID, rxID, TransitionApproaching); err != nil {
			return fmt.Errorf("creating notification for prescription %d: %w", rxID, err)
		}
	}
	return nil
}

// List returns all notifications for a pharmacy.
func (s *Service) List(ctx context.Context, pharmacyID int64) ([]Notification, error) {
	notifs, err := s.deps.Lister.ListByPharmacy(ctx, pharmacyID)
	if err != nil {
		return nil, fmt.Errorf("listing notifications: %w", err)
	}
	return notifs, nil
}

// MarkRead marks a single notification as read.
func (s *Service) MarkRead(ctx context.Context, id, pharmacyID int64) error {
	if err := s.deps.Reader.MarkRead(ctx, id, pharmacyID); err != nil {
		return fmt.Errorf("marking notification read: %w", err)
	}
	return nil
}

// MarkAllRead marks all notifications as read for a pharmacy.
func (s *Service) MarkAllRead(ctx context.Context, pharmacyID int64) error {
	if err := s.deps.AllReader.MarkAllRead(ctx, pharmacyID); err != nil {
		return fmt.Errorf("marking all notifications read: %w", err)
	}
	return nil
}

// CountUnread returns the number of unread notifications.
func (s *Service) CountUnread(ctx context.Context, pharmacyID int64) (int64, error) {
	count, err := s.deps.Counter.CountUnread(ctx, pharmacyID)
	if err != nil {
		return 0, fmt.Errorf("counting unread notifications: %w", err)
	}
	return count, nil
}
