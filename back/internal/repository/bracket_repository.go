package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type BracketRepository struct {
	db Queryer
}

func NewBracketRepository(db Queryer) *BracketRepository {
	return &BracketRepository{db: db}
}

func (r *BracketRepository) CreateBracket(ctx context.Context, b *entity.Bracket) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO brackets (id, tournament_id, format, seeding_method, status, generated_by, generated_at, metadata_json)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
    `, b.ID, b.TournamentID, b.Format, b.SeedingMethod, b.Status, b.GeneratedBy, b.GeneratedAt, b.MetadataJSON)
	return err
}

func (r *BracketRepository) UpdateBracketStatus(ctx context.Context, bracketID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE brackets SET status=$2 WHERE id=$1`, bracketID, status)
	return err
}

func (r *BracketRepository) GetByTournamentID(ctx context.Context, tournamentID string) (*entity.Bracket, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, tournament_id, format, seeding_method, status, generated_by, generated_at, metadata_json
        FROM brackets WHERE tournament_id=$1
    `, tournamentID)
	var b entity.Bracket
	err := row.Scan(&b.ID, &b.TournamentID, &b.Format, &b.SeedingMethod, &b.Status, &b.GeneratedBy, &b.GeneratedAt, &b.MetadataJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &b, nil
}

func (r *BracketRepository) DeleteByTournamentID(ctx context.Context, tournamentID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM matches WHERE tournament_id=$1`, tournamentID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM brackets WHERE tournament_id=$1`, tournamentID)
	return err
}

func (r *BracketRepository) CreateMatch(ctx context.Context, m *entity.Match) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO matches (
            id, tournament_id, bracket_id, round_number, slot_index, team1_id, team2_id, scheduled_at, location_or_server,
            status, team1_confirmation_status, team2_confirmation_status, winner_team_id, score_text, manager_comment,
            next_match_id, source_match1_id, source_match2_id, is_bye, deleted_at
        )
        VALUES (
            $1,$2,$3,$4,$5,$6,$7,$8,$9,
            $10,$11,$12,$13,$14,$15,
            $16,$17,$18,$19,$20
        )
    `, m.ID, m.TournamentID, m.BracketID, m.RoundNumber, m.SlotIndex, m.Team1ID, m.Team2ID, m.ScheduledAt, m.LocationOrServer,
		m.Status, m.Team1ConfirmationStatus, m.Team2ConfirmationStatus, m.WinnerTeamID, m.ScoreText, m.ManagerComment,
		m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.IsBye, m.DeletedAt)
	return err
}

func (r *BracketRepository) UpdateNextMatchID(ctx context.Context, matchID string, nextMatchID *string) error {
	_, err := r.db.Exec(ctx, `UPDATE matches SET next_match_id=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, matchID, nextMatchID)
	return err
}

func (r *BracketRepository) UpdateMatchState(ctx context.Context, m *entity.Match) error {
	_, err := r.db.Exec(ctx, `
        UPDATE matches
        SET team1_id=$2, team2_id=$3, scheduled_at=$4, location_or_server=$5, status=$6,
            team1_confirmation_status=$7, team2_confirmation_status=$8, winner_team_id=$9, score_text=$10,
            manager_comment=$11, next_match_id=$12, source_match1_id=$13, source_match2_id=$14, is_bye=$15, updated_at=now()
        WHERE id=$1 AND deleted_at IS NULL
    `, m.ID, m.Team1ID, m.Team2ID, m.ScheduledAt, m.LocationOrServer, m.Status, m.Team1ConfirmationStatus, m.Team2ConfirmationStatus, m.WinnerTeamID, m.ScoreText, m.ManagerComment, m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID, m.IsBye)
	return err
}

func (r *BracketRepository) GetMatchByID(ctx context.Context, matchID string) (*entity.Match, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, tournament_id, bracket_id, round_number, slot_index, team1_id, team2_id, scheduled_at, location_or_server,
               status, team1_confirmation_status, team2_confirmation_status, winner_team_id, score_text, manager_comment,
               next_match_id, source_match1_id, source_match2_id, is_bye, created_at, updated_at, deleted_at
        FROM matches WHERE id=$1 AND deleted_at IS NULL
    `, matchID)
	return scanMatch(row)
}

func (r *BracketRepository) ListMatchesByTournament(ctx context.Context, tournamentID string) ([]entity.Match, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, tournament_id, bracket_id, round_number, slot_index, team1_id, team2_id, scheduled_at, location_or_server,
               status, team1_confirmation_status, team2_confirmation_status, winner_team_id, score_text, manager_comment,
               next_match_id, source_match1_id, source_match2_id, is_bye, created_at, updated_at, deleted_at
        FROM matches
        WHERE tournament_id=$1 AND deleted_at IS NULL
        ORDER BY round_number ASC, slot_index ASC
    `, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Match, 0)
	for rows.Next() {
		match, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *match)
	}
	return result, rows.Err()
}

func scanMatch(row interface {
	Scan(dest ...interface{}) error
}) (*entity.Match, error) {
	var m entity.Match
	err := row.Scan(
		&m.ID, &m.TournamentID, &m.BracketID, &m.RoundNumber, &m.SlotIndex, &m.Team1ID, &m.Team2ID, &m.ScheduledAt, &m.LocationOrServer,
		&m.Status, &m.Team1ConfirmationStatus, &m.Team2ConfirmationStatus, &m.WinnerTeamID, &m.ScoreText, &m.ManagerComment,
		&m.NextMatchID, &m.SourceMatch1ID, &m.SourceMatch2ID, &m.IsBye, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &m, nil
}
