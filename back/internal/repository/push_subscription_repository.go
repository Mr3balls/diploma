package repository

import (
	"context"

	"esports-backend/internal/entity"
)

type PushSubscriptionRepository struct {
	db Queryer
}

func NewPushSubscriptionRepository(db Queryer) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db}
}

func (r *PushSubscriptionRepository) Save(ctx context.Context, sub *entity.PushSubscription) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, endpoint) DO UPDATE SET p256dh=$3, auth=$4
    `, sub.UserID, sub.Endpoint, sub.P256dh, sub.Auth)
	return err
}

func (r *PushSubscriptionRepository) Delete(ctx context.Context, userID, endpoint string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM push_subscriptions WHERE user_id=$1 AND endpoint=$2`,
		userID, endpoint,
	)
	return err
}

func (r *PushSubscriptionRepository) ListByUserID(ctx context.Context, userID string) ([]entity.PushSubscription, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, endpoint, p256dh, auth, created_at FROM push_subscriptions WHERE user_id=$1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []entity.PushSubscription
	for rows.Next() {
		var s entity.PushSubscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.Endpoint, &s.P256dh, &s.Auth, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// DeleteByEndpoint removes a push subscription by endpoint across all users (for expired sub cleanup).
func (r *PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint=$1`, endpoint)
	return err
}
