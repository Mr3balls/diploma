package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type ParticipantRepository struct {
	db Queryer
}

func NewParticipantRepository(db Queryer) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

func (r *ParticipantRepository) Create(ctx context.Context, p *entity.Participant) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO participants (id, tournament_id, user_id, name, seed, status, final_rank, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, p.ID, p.TournamentID, p.UserID, p.Name, p.Seed, p.Status, p.FinalRank, p.CreatedAt)
	return err
}

func (r *ParticipantRepository) GetByID(ctx context.Context, id string) (*entity.Participant, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, tournament_id, user_id, name, seed, status, final_rank, created_at
		FROM participants WHERE id=$1
	`, id)
	return scanParticipant(row)
}

func (r *ParticipantRepository) GetByUserAndTournament(ctx context.Context, tournamentID, userID string) (*entity.Participant, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, tournament_id, user_id, name, seed, status, final_rank, created_at
		FROM participants WHERE tournament_id=$1 AND user_id=$2
	`, tournamentID, userID)
	return scanParticipant(row)
}

func (r *ParticipantRepository) ListByTournament(ctx context.Context, tournamentID string) ([]*entity.Participant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, tournament_id, user_id, name, seed, status, final_rank, created_at
		FROM participants WHERE tournament_id=$1
		ORDER BY seed ASC, created_at ASC
	`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*entity.Participant
	for rows.Next() {
		p, err := scanParticipant(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ParticipantRepository) Count(ctx context.Context, tournamentID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM participants WHERE tournament_id=$1`, tournamentID).Scan(&count)
	return count, err
}

func (r *ParticipantRepository) UpdateSeed(ctx context.Context, id string, seed int) error {
	_, err := r.db.Exec(ctx, `UPDATE participants SET seed=$2 WHERE id=$1`, id, seed)
	return err
}

func (r *ParticipantRepository) UpdateStatus(ctx context.Context, id, status string, finalRank *int) error {
	_, err := r.db.Exec(ctx, `UPDATE participants SET status=$2, final_rank=$3 WHERE id=$1`, id, status, finalRank)
	return err
}

func (r *ParticipantRepository) LinkUser(ctx context.Context, id, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE participants SET user_id=$2 WHERE id=$1`, id, userID)
	return err
}

func (r *ParticipantRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM participants WHERE id=$1`, id)
	return err
}

func (r *ParticipantRepository) DeleteByTournament(ctx context.Context, tournamentID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM participants WHERE tournament_id=$1`, tournamentID)
	return err
}

func scanParticipant(row interface{ Scan(...interface{}) error }) (*entity.Participant, error) {
	var p entity.Participant
	err := row.Scan(&p.ID, &p.TournamentID, &p.UserID, &p.Name, &p.Seed, &p.Status, &p.FinalRank, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &p, nil
}
