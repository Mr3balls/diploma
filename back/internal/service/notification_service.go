package service

import (
	"context"

	"esports-backend/internal/repository"
)

type NotificationService struct {
	notifications *repository.NotificationRepository
}

func NewNotificationService(notifications *repository.NotificationRepository) *NotificationService {
	return &NotificationService{notifications: notifications}
}

func (s *NotificationService) List(ctx context.Context, userID string, limit, offset int) (interface{}, error) {
	return s.notifications.ListByUser(ctx, userID, limit, offset)
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID string) (int, error) {
	return s.notifications.GetUnreadCount(ctx, userID)
}

func (s *NotificationService) Read(ctx context.Context, userID, notificationID string) error {
	return s.notifications.MarkRead(ctx, notificationID, userID)
}

func (s *NotificationService) ReadAll(ctx context.Context, userID string) error {
	return s.notifications.MarkAllRead(ctx, userID)
}

func (s *NotificationService) Act(ctx context.Context, userID, notificationID string) error {
	return s.notifications.MarkActed(ctx, notificationID, userID)
}
