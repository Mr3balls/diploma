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
	// Null out self-referential FKs first to avoid constraint violations on delete.
	_, err := r.db.Exec(ctx, `
		UPDATE matches
		SET next_match_id=NULL, source_match1_id=NULL, source_match2_id=NULL, loser_next_match_id=NULL
		WHERE tournament_id=$1
	`, tournamentID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM matches WHERE tournament_id=$1`, tournamentID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `DELETE FROM brackets WHERE tournament_id=$1`, tournamentID)
	return err
}

func (r *BracketRepository) CreateMatch(ctx context.Context, m *entity.Match) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO matches (
            id, tournament_id, bracket_id, group_id, bracket_section, round_number, slot_index,
            global_number,
            team1_id, team2_id, participant1_id, participant2_id,
            scheduled_at, location_or_server,
            status, team1_confirmation_status, team2_confirmation_status,
            winner_team_id, winner_participant_id, score_text, manager_comment,
            next_match_id, source_match1_id, source_match2_id,
            loser_next_match_id, loser_next_slot,
            is_bye, deleted_at
        ) VALUES (
            $1,$2,$3,$4,$5,$6,$7,
            $8,
            $9,$10,$11,$12,
            $13,$14,
            $15,$16,$17,
            $18,$19,$20,$21,
            $22,$23,$24,
            $25,$26,
            $27,$28
        )
    `,
		m.ID, m.TournamentID, m.BracketID, m.GroupID, m.BracketSection, m.RoundNumber, m.SlotIndex,
		m.GlobalNumber,
		m.Team1ID, m.Team2ID, m.Participant1ID, m.Participant2ID,
		m.ScheduledAt, m.LocationOrServer,
		m.Status, m.Team1ConfirmationStatus, m.Team2ConfirmationStatus,
		m.WinnerTeamID, m.WinnerParticipantID, m.ScoreText, m.ManagerComment,
		m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID,
		m.LoserNextMatchID, m.LoserNextSlot,
		m.IsBye, m.DeletedAt,
	)
	return err
}

func (r *BracketRepository) DeleteMatchByID(ctx context.Context, matchID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM matches WHERE id=$1`, matchID)
	return err
}

func (r *BracketRepository) UpdateNextMatchID(ctx context.Context, matchID string, nextMatchID *string) error {
	_, err := r.db.Exec(ctx, `UPDATE matches SET next_match_id=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, matchID, nextMatchID)
	return err
}

func (r *BracketRepository) UpdateMatchState(ctx context.Context, m *entity.Match) error {
	_, err := r.db.Exec(ctx, `
        UPDATE matches
        SET team1_id=$2, team2_id=$3, participant1_id=$4, participant2_id=$5,
            scheduled_at=$6, location_or_server=$7, status=$8,
            team1_confirmation_status=$9, team2_confirmation_status=$10,
            winner_team_id=$11, winner_participant_id=$12,
            score_text=$13, manager_comment=$14,
            next_match_id=$15, source_match1_id=$16, source_match2_id=$17,
            loser_next_match_id=$18, loser_next_slot=$19,
            is_bye=$20, updated_at=now()
        WHERE id=$1 AND deleted_at IS NULL
    `,
		m.ID,
		m.Team1ID, m.Team2ID, m.Participant1ID, m.Participant2ID,
		m.ScheduledAt, m.LocationOrServer, m.Status,
		m.Team1ConfirmationStatus, m.Team2ConfirmationStatus,
		m.WinnerTeamID, m.WinnerParticipantID,
		m.ScoreText, m.ManagerComment,
		m.NextMatchID, m.SourceMatch1ID, m.SourceMatch2ID,
		m.LoserNextMatchID, m.LoserNextSlot,
		m.IsBye,
	)
	return err
}

func (r *BracketRepository) GetMatchByID(ctx context.Context, matchID string) (*entity.Match, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, tournament_id, bracket_id, group_id, bracket_section, round_number, slot_index,
               COALESCE(global_number,0),
               team1_id, team2_id, participant1_id, participant2_id,
               scheduled_at, location_or_server,
               status, team1_confirmation_status, team2_confirmation_status,
               winner_team_id, winner_participant_id, score_text, manager_comment,
               next_match_id, source_match1_id, source_match2_id,
               loser_next_match_id, loser_next_slot,
               is_bye, created_at, updated_at, deleted_at
        FROM matches WHERE id=$1 AND deleted_at IS NULL
    `, matchID)
	return scanMatch(row)
}

func (r *BracketRepository) ListMatchesByTournament(ctx context.Context, tournamentID string) ([]entity.Match, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, tournament_id, bracket_id, group_id, bracket_section, round_number, slot_index,
               COALESCE(global_number,0),
               team1_id, team2_id, participant1_id, participant2_id,
               scheduled_at, location_or_server,
               status, team1_confirmation_status, team2_confirmation_status,
               winner_team_id, winner_participant_id, score_text, manager_comment,
               next_match_id, source_match1_id, source_match2_id,
               loser_next_match_id, loser_next_slot,
               is_bye, created_at, updated_at, deleted_at
        FROM matches
        WHERE tournament_id=$1 AND deleted_at IS NULL
        ORDER BY
            group_id NULLS LAST,
            CASE bracket_section WHEN 'WB' THEN 0 WHEN 'LB' THEN 1 WHEN 'GF' THEN 2 END,
            round_number ASC, slot_index ASC
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
		&m.ID, &m.TournamentID, &m.BracketID, &m.GroupID, &m.BracketSection, &m.RoundNumber, &m.SlotIndex,
		&m.GlobalNumber,
		&m.Team1ID, &m.Team2ID, &m.Participant1ID, &m.Participant2ID,
		&m.ScheduledAt, &m.LocationOrServer,
		&m.Status, &m.Team1ConfirmationStatus, &m.Team2ConfirmationStatus,
		&m.WinnerTeamID, &m.WinnerParticipantID, &m.ScoreText, &m.ManagerComment,
		&m.NextMatchID, &m.SourceMatch1ID, &m.SourceMatch2ID,
		&m.LoserNextMatchID, &m.LoserNextSlot,
		&m.IsBye, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &m, nil
}
