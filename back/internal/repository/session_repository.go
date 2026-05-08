package repository

import (
	"context"
	"errors"
	"time"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type SessionRepository struct {
	db Queryer
}

func NewSessionRepository(db Queryer) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *entity.AuthSession) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO auth_sessions (id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7)
    `, s.ID, s.UserID, s.RefreshTokenHash, s.UserAgent, s.IPAddress, s.ExpiresAt, s.RevokedAt)
	return err
}

func (r *SessionRepository) GetActiveByHash(ctx context.Context, hash string) (*entity.AuthSession, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at
        FROM auth_sessions
        WHERE refresh_token_hash = $1 AND revoked_at IS NULL AND expires_at > now()
    `, hash)
	var s entity.AuthSession
	err := row.Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress, &s.ExpiresAt, &s.RevokedAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepository) RevokeByHash(ctx context.Context, hash string) error {
	_, err := r.db.Exec(ctx, `UPDATE auth_sessions SET revoked_at=now() WHERE refresh_token_hash=$1 AND revoked_at IS NULL`, hash)
	return err
}

func (r *SessionRepository) RevokeByUserID(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE auth_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	return err
}

func (r *SessionRepository) CleanupExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `DELETE FROM auth_sessions WHERE revoked_at IS NOT NULL OR expires_at < $1`, time.Now())
	return err
}
