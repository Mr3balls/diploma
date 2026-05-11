package repository

import (
	"context"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type MatchLogRepository struct {
	db Queryer
}

func NewMatchLogRepository(db Queryer) *MatchLogRepository {
	return &MatchLogRepository{db: db}
}

func (r *MatchLogRepository) Create(ctx context.Context, e *entity.MatchLogEntry) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO match_log_entries (id, tournament_id, match_id, action, actor_id, detail, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, e.ID, e.TournamentID, e.MatchID, e.Action, e.ActorID, e.Detail, e.CreatedAt)
	return err
}

func (r *MatchLogRepository) ListByTournament(ctx context.Context, tournamentID string) ([]*entity.MatchLogEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tournament_id, match_id, action, actor_id, detail, created_at
		FROM match_log_entries WHERE tournament_id=$1
		ORDER BY created_at DESC
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatchLogEntries(rows)
}

func (r *MatchLogRepository) ListByMatch(ctx context.Context, matchID string) ([]*entity.MatchLogEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tournament_id, match_id, action, actor_id, detail, created_at
		FROM match_log_entries WHERE match_id=$1
		ORDER BY created_at DESC
	`, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatchLogEntries(rows)
}

func scanMatchLogEntries(rows pgx.Rows) ([]*entity.MatchLogEntry, error) {
	var out []*entity.MatchLogEntry
	for rows.Next() {
		var e entity.MatchLogEntry
		if err := rows.Scan(&e.ID, &e.TournamentID, &e.MatchID, &e.Action, &e.ActorID, &e.Detail, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}
