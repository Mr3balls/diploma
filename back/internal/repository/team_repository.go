package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type TeamRepository struct {
	db Queryer
}

func NewTeamRepository(db Queryer) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, t *entity.Team) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO teams (id, tournament_id, name, status, approved_by_manager, created_from_import_row_id)
        VALUES ($1,$2,$3,$4,$5,$6)
    `, t.ID, t.TournamentID, t.Name, t.Status, t.ApprovedByManager, t.CreatedFromImportRowID)
	return err
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, t *entity.Team) error {
	_, err := r.db.Exec(ctx, `UPDATE teams SET name=$2, status=$3, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, t.ID, t.Name, t.Status)
	return err
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id string) (*entity.Team, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, tournament_id, name, status, approved_by_manager, created_from_import_row_id, created_at, updated_at, deleted_at
        FROM teams WHERE id=$1 AND deleted_at IS NULL
    `, id)
	var t entity.Team
	err := row.Scan(&t.ID, &t.TournamentID, &t.Name, &t.Status, &t.ApprovedByManager, &t.CreatedFromImportRowID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &t, nil
}

func (r *TeamRepository) CountByTournament(ctx context.Context, tournamentID string) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM teams WHERE tournament_id=$1 AND deleted_at IS NULL AND status != 'rejected'`,
		tournamentID,
	).Scan(&n)
	return n, err
}

func (r *TeamRepository) ListByTournament(ctx context.Context, tournamentID string, admin bool) ([]entity.Team, error) {
	query := `
        SELECT id, tournament_id, name, status, approved_by_manager, created_from_import_row_id, created_at, updated_at, deleted_at
        FROM teams
        WHERE tournament_id=$1 AND deleted_at IS NULL
    `
	if !admin {
		query += ` AND status != 'rejected'`
	}
	query += ` ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Team, 0)
	for rows.Next() {
		var t entity.Team
		if err := rows.Scan(&t.ID, &t.TournamentID, &t.Name, &t.Status, &t.ApprovedByManager, &t.CreatedFromImportRowID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *TeamRepository) ListApprovedTeamIDs(ctx context.Context, tournamentID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT id FROM teams WHERE tournament_id=$1 AND status='approved' AND deleted_at IS NULL ORDER BY created_at ASC`, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, rows.Err()
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE teams SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *TeamRepository) SetApproval(ctx context.Context, id, status string, approved *bool) error {
	_, err := r.db.Exec(ctx, `UPDATE teams SET status=$2, approved_by_manager=$3, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id, status, approved)
	return err
}

func (r *TeamRepository) CreateMember(ctx context.Context, m *entity.TeamMember) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO team_members (id, team_id, user_id, nickname, email, member_role, is_captain, is_substitute, confirmation_status, invited_at, responded_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
    `, m.ID, m.TeamID, m.UserID, m.Nickname, m.Email, m.MemberRole, m.IsCaptain, m.IsSubstitute, m.ConfirmationStatus, m.InvitedAt, m.RespondedAt)
	return err
}

func (r *TeamRepository) GetMemberByID(ctx context.Context, id string) (*entity.TeamMember, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, team_id, user_id, nickname, email, member_role, is_captain, is_substitute, confirmation_status, invited_at, responded_at, created_at, updated_at, deleted_at
        FROM team_members WHERE id=$1 AND deleted_at IS NULL
    `, id)
	var m entity.TeamMember
	err := row.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Nickname, &m.Email, &m.MemberRole, &m.IsCaptain, &m.IsSubstitute, &m.ConfirmationStatus, &m.InvitedAt, &m.RespondedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &m, nil
}

func (r *TeamRepository) ListMembersByTeamID(ctx context.Context, teamID string) ([]entity.TeamMember, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, team_id, user_id, nickname, email, member_role, is_captain, is_substitute, confirmation_status, invited_at, responded_at, created_at, updated_at, deleted_at
        FROM team_members WHERE team_id=$1 AND deleted_at IS NULL
        ORDER BY is_captain DESC, is_substitute ASC, created_at ASC
    `, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.TeamMember, 0)
	for rows.Next() {
		var m entity.TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Nickname, &m.Email, &m.MemberRole, &m.IsCaptain, &m.IsSubstitute, &m.ConfirmationStatus, &m.InvitedAt, &m.RespondedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *TeamRepository) SetMemberConfirmation(ctx context.Context, memberID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE team_members SET confirmation_status=$2, responded_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, memberID, status)
	return err
}

func (r *TeamRepository) RemoveMember(ctx context.Context, memberID string) error {
	_, err := r.db.Exec(ctx, `UPDATE team_members SET confirmation_status='removed', deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, memberID)
	return err
}

func (r *TeamRepository) ReplaceMember(ctx context.Context, memberID string, userID *string, nickname string, email *string) error {
	_, err := r.db.Exec(ctx, `UPDATE team_members SET user_id=$2, nickname=$3, email=$4, confirmation_status='pending_confirmation', responded_at=NULL, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, memberID, userID, nickname, email)
	return err
}

func (r *TeamRepository) FindPendingByEmail(ctx context.Context, email string) ([]entity.TeamMember, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, team_id, user_id, nickname, email, member_role, is_captain, is_substitute, confirmation_status, invited_at, responded_at, created_at, updated_at, deleted_at
        FROM team_members WHERE email=$1 AND user_id IS NULL AND deleted_at IS NULL
    `, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.TeamMember, 0)
	for rows.Next() {
		var m entity.TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Nickname, &m.Email, &m.MemberRole, &m.IsCaptain, &m.IsSubstitute, &m.ConfirmationStatus, &m.InvitedAt, &m.RespondedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *TeamRepository) SetMemberUserID(ctx context.Context, memberID, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE team_members SET user_id=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, memberID, userID)
	return err
}

func (r *TeamRepository) FindCaptainMembership(ctx context.Context, userID, tournamentID string) ([]entity.TeamMember, error) {
	rows, err := r.db.Query(ctx, `
        SELECT tm.id, tm.team_id, tm.user_id, tm.nickname, tm.email, tm.member_role, tm.is_captain, tm.is_substitute, tm.confirmation_status, tm.invited_at, tm.responded_at, tm.created_at, tm.updated_at, tm.deleted_at
        FROM team_members tm
        JOIN teams t ON t.id = tm.team_id
        WHERE tm.user_id=$1 AND tm.is_captain=true AND t.tournament_id=$2 AND tm.deleted_at IS NULL AND t.deleted_at IS NULL
    `, userID, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.TeamMember, 0)
	for rows.Next() {
		var m entity.TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.Nickname, &m.Email, &m.MemberRole, &m.IsCaptain, &m.IsSubstitute, &m.ConfirmationStatus, &m.InvitedAt, &m.RespondedAt, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, rows.Err()
}
