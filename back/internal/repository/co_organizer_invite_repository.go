package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type CoOrganizerInviteRepository struct {
	db Queryer
}

func NewCoOrganizerInviteRepository(db Queryer) *CoOrganizerInviteRepository {
	return &CoOrganizerInviteRepository{db: db}
}

func (r *CoOrganizerInviteRepository) Create(ctx context.Context, inv *entity.CoOrganizerInvite) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO co_organizer_invites (id, tournament_id, invitee_id, invited_by_id, token, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, inv.ID, inv.TournamentID, inv.InviteeID, inv.InvitedByID, inv.Token, inv.Status, inv.ExpiresAt, inv.CreatedAt)
	return err
}

func (r *CoOrganizerInviteRepository) GetByToken(ctx context.Context, token string) (*entity.CoOrganizerInvite, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, tournament_id, invitee_id, invited_by_id, token, status, expires_at, created_at
		FROM co_organizer_invites WHERE token=$1
	`, token)
	return scanInvite(row)
}

func (r *CoOrganizerInviteRepository) GetByID(ctx context.Context, id string) (*entity.CoOrganizerInvite, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, tournament_id, invitee_id, invited_by_id, token, status, expires_at, created_at
		FROM co_organizer_invites WHERE id=$1
	`, id)
	return scanInvite(row)
}

func (r *CoOrganizerInviteRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE co_organizer_invites SET status=$2 WHERE id=$1`, id, status)
	return err
}

// MarkExpired transitions all pending invites past their expiry to 'expired'.
func (r *CoOrganizerInviteRepository) MarkExpired(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		UPDATE co_organizer_invites SET status='expired'
		WHERE status='pending' AND expires_at < NOW()
	`)
	return err
}

func scanInvite(row interface{ Scan(...interface{}) error }) (*entity.CoOrganizerInvite, error) {
	var inv entity.CoOrganizerInvite
	err := row.Scan(&inv.ID, &inv.TournamentID, &inv.InviteeID, &inv.InvitedByID,
		&inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}
	return &inv, err
}
