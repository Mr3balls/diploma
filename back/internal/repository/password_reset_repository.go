package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type PasswordResetRepository struct {
	db Queryer
}

func NewPasswordResetRepository(db Queryer) *PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(ctx context.Context, t *entity.PasswordResetToken) error {
	// Invalidate any existing unused tokens for this user first.
	_, _ = r.db.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at = now() WHERE user_id = $1 AND used_at IS NULL`,
		t.UserID,
	)
	_, err := r.db.Exec(ctx,
		`INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at) VALUES ($1,$2,$3,$4)`,
		t.ID, t.UserID, t.TokenHash, t.ExpiresAt,
	)
	return err
}

func (r *PasswordResetRepository) GetActiveByHash(ctx context.Context, hash string) (*entity.PasswordResetToken, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		 FROM password_reset_tokens
		 WHERE token_hash = $1 AND used_at IS NULL AND expires_at > now()`,
		hash,
	)
	var t entity.PasswordResetToken
	err := row.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &t, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at = now() WHERE id = $1`,
		id,
	)
	return err
}
