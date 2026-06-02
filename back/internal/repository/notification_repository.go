package repository

import (
	"context"
	"encoding/json"

	"esports-backend/internal/entity"
)

// NotificationBroadcaster is implemented by the SSE hub to push real-time events.
type NotificationBroadcaster interface {
	BroadcastToUser(userID string)
}

// PushSender sends Web Push notifications to a user's registered devices.
type PushSender interface {
	SendToUser(userID, title, body string)
}

type NotificationRepository struct {
	db          Queryer
	broadcaster NotificationBroadcaster
	pushSender  PushSender
}

func NewNotificationRepository(db Queryer) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) WithBroadcaster(b NotificationBroadcaster) *NotificationRepository {
	r.broadcaster = b
	return r
}

func (r *NotificationRepository) WithPushSender(p PushSender) *NotificationRepository {
	r.pushSender = p
	return r
}

func (r *NotificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	// Check notification_preferences: skip if this type is disabled for the user.
	if r.isTypeDisabled(ctx, n.UserID, n.Type) {
		return nil
	}

	_, err := r.db.Exec(ctx, `
        INSERT INTO notifications (id, user_id, type, title, message, payload_json, action_payload_json, is_read, acted_at, read_at, deleted_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
    `, n.ID, n.UserID, n.Type, n.Title, n.Message, n.PayloadJSON, n.ActionPayloadJSON, n.IsRead, n.ActedAt, n.ReadAt, n.DeletedAt)
	if err != nil {
		return err
	}

	if r.broadcaster != nil {
		r.broadcaster.BroadcastToUser(n.UserID)
	}
	if r.pushSender != nil {
		go r.pushSender.SendToUser(n.UserID, n.Title, n.Message)
	}
	return nil
}

// isTypeDisabled checks the user's notification_preferences and returns true if the type is in the disabled list.
func (r *NotificationRepository) isTypeDisabled(ctx context.Context, userID, notifType string) bool {
	var raw []byte
	err := r.db.QueryRow(ctx,
		`SELECT notification_preferences FROM users WHERE id=$1 AND deleted_at IS NULL`,
		userID,
	).Scan(&raw)
	if err != nil || raw == nil {
		return false
	}
	var p struct {
		Disabled []string `json:"disabled"`
	}
	if json.Unmarshal(raw, &p) != nil {
		return false
	}
	for _, d := range p.Disabled {
		if d == notifType {
			return true
		}
	}
	return false
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]entity.Notification, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, type, title, message, payload_json, action_payload_json, is_read, acted_at, read_at, created_at, deleted_at
        FROM notifications
        WHERE user_id=$1 AND deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Notification, 0)
	for rows.Next() {
		var n entity.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.PayloadJSON, &n.ActionPayloadJSON, &n.IsRead, &n.ActedAt, &n.ReadAt, &n.CreatedAt, &n.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND is_read=false AND deleted_at IS NULL`, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, notificationID, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET is_read=true, read_at=now() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, notificationID, userID)
	return err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET is_read=true, read_at=now() WHERE user_id=$1 AND is_read=false AND deleted_at IS NULL`, userID)
	return err
}

func (r *NotificationRepository) MarkActed(ctx context.Context, notificationID, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET acted_at=now(), is_read=true, read_at=now() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, notificationID, userID)
	return err
}

// SoftDelete soft-deletes a single notification belonging to userID.
func (r *NotificationRepository) SoftDelete(ctx context.Context, notificationID, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET deleted_at=now() WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL`, notificationID, userID)
	return err
}

// SoftDeleteAll soft-deletes all notifications for a user.
func (r *NotificationRepository) SoftDeleteAll(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET deleted_at=now() WHERE user_id=$1 AND deleted_at IS NULL`, userID)
	return err
}
