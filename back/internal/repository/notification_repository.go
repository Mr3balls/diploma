package repository

import (
	"context"

	"esports-backend/internal/entity"
)

type NotificationRepository struct {
	db Queryer
}

func NewNotificationRepository(db Queryer) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *entity.Notification) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO notifications (id, user_id, type, title, message, payload_json, action_payload_json, is_read, acted_at, read_at, deleted_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
    `, n.ID, n.UserID, n.Type, n.Title, n.Message, n.PayloadJSON, n.ActionPayloadJSON, n.IsRead, n.ActedAt, n.ReadAt, n.DeletedAt)
	return err
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
