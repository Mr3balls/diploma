package service

import (
	"context"

	"esports-backend/internal/entity"
	"esports-backend/internal/repository"
)

type NotificationService struct {
	notifications *repository.NotificationRepository
	users         *repository.UserRepository
	push          *PushService
}

func NewNotificationService(notifications *repository.NotificationRepository, users *repository.UserRepository, push *PushService) *NotificationService {
	return &NotificationService{notifications: notifications, users: users, push: push}
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

func (s *NotificationService) Delete(ctx context.Context, userID, notificationID string) error {
	return s.notifications.SoftDelete(ctx, notificationID, userID)
}

func (s *NotificationService) DeleteAll(ctx context.Context, userID string) error {
	return s.notifications.SoftDeleteAll(ctx, userID)
}

func (s *NotificationService) GetPreferences(ctx context.Context, userID string) ([]string, error) {
	disabled, err := s.users.GetNotificationPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}
	if disabled == nil {
		disabled = []string{}
	}
	return disabled, nil
}

func (s *NotificationService) SetPreferences(ctx context.Context, userID string, disabled []string) error {
	return s.users.SetNotificationPreferences(ctx, userID, disabled)
}

func (s *NotificationService) VAPIDPublicKey() string {
	if s.push == nil {
		return ""
	}
	return s.push.VAPIDPublicKey()
}

func (s *NotificationService) RegisterPush(ctx context.Context, userID, endpoint, p256dh, auth string) error {
	if s.push == nil {
		return nil
	}
	return s.push.subs.Save(ctx, &entity.PushSubscription{UserID: userID, Endpoint: endpoint, P256dh: p256dh, Auth: auth})
}

func (s *NotificationService) UnregisterPush(ctx context.Context, userID, endpoint string) error {
	if s.push == nil {
		return nil
	}
	return s.push.subs.Delete(ctx, userID, endpoint)
}
