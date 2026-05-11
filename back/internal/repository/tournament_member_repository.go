package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type TournamentMemberRepository struct {
	db Queryer
}

func NewTournamentMemberRepository(db Queryer) *TournamentMemberRepository {
	return &TournamentMemberRepository{db: db}
}

func (r *TournamentMemberRepository) Upsert(ctx context.Context, m *entity.TournamentMember) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO tournament_members (id, tournament_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tournament_id, user_id) DO UPDATE SET role=EXCLUDED.role
	`, m.ID, m.TournamentID, m.UserID, m.Role, m.JoinedAt)
	return err
}

func (r *TournamentMemberRepository) GetRole(ctx context.Context, tournamentID, userID string) (string, error) {
	var role string
	err := r.db.QueryRow(ctx, `
		SELECT role FROM tournament_members WHERE tournament_id=$1 AND user_id=$2
	`, tournamentID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", pgx.ErrNoRows
	}
	return role, err
}

func (r *TournamentMemberRepository) List(ctx context.Context, tournamentID string) ([]*entity.TournamentMember, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tournament_id, user_id, role, joined_at
		FROM tournament_members WHERE tournament_id=$1
		ORDER BY joined_at ASC
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*entity.TournamentMember
	for rows.Next() {
		var m entity.TournamentMember
		if err := rows.Scan(&m.ID, &m.TournamentID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

func (r *TournamentMemberRepository) Delete(ctx context.Context, tournamentID, userID string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM tournament_members WHERE tournament_id=$1 AND user_id=$2
	`, tournamentID, userID)
	return err
}

func (r *TournamentMemberRepository) DeleteByTournament(ctx context.Context, tournamentID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tournament_members WHERE tournament_id=$1`, tournamentID)
	return err
}
