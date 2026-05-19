package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

// TournamentFilter holds optional filter parameters for public tournament queries.
type TournamentFilter struct {
	Status      string // exact match, e.g. "registration_open"
	Format      string // exact match, e.g. "single_elimination"
	Discipline  string // case-insensitive substring match
	Query       string // searches title and discipline via ILIKE
	RequesterID string // if set, also includes private tournaments where user has a role
}

// publicWhere builds the WHERE clause and args for public tournament queries.
// Returns (whereClause, args) where whereClause starts with "WHERE ".
func publicWhere(f TournamentFilter) (string, []interface{}) {
	args := []interface{}{}

	// visibility condition: always show public; also show private ones the requester manages
	var visibilityCond string
	if f.RequesterID != "" {
		args = append(args, f.RequesterID)
		visibilityCond = fmt.Sprintf(
			"(visibility='public' OR (visibility='private' AND id IN (SELECT tournament_id FROM tournament_user_roles WHERE user_id=$%d)))",
			len(args),
		)
	} else {
		visibilityCond = "visibility='public'"
	}

	conds := []string{"deleted_at IS NULL", visibilityCond}

	if f.Status != "" {
		args = append(args, f.Status)
		conds = append(conds, fmt.Sprintf("status=$%d", len(args)))
	}
	if f.Format != "" {
		args = append(args, f.Format)
		conds = append(conds, fmt.Sprintf("format=$%d", len(args)))
	}
	if f.Discipline != "" {
		args = append(args, "%"+f.Discipline+"%")
		conds = append(conds, fmt.Sprintf("discipline ILIKE $%d", len(args)))
	}
	if f.Query != "" {
		args = append(args, "%"+f.Query+"%")
		conds = append(conds, fmt.Sprintf("(title ILIKE $%d OR discipline ILIKE $%d)", len(args), len(args)))
	}

	return "WHERE " + strings.Join(conds, " AND "), args
}

type TournamentRepository struct {
	db Queryer
}

func NewTournamentRepository(db Queryer) *TournamentRepository {
	return &TournamentRepository{db: db}
}

func (r *TournamentRepository) Create(ctx context.Context, t *entity.Tournament) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO tournaments (id, title, discipline, description, rules, location, max_teams, format, group_count, registration_deadline, start_at, status, visibility, slug, max_participants, owner_user_id, registration_mode, latitude, longitude)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
    `, t.ID, t.Title, t.Discipline, t.Description, t.Rules, t.Location, t.MaxTeams, t.Format, t.GroupCount, t.RegistrationDeadline, t.StartAt, t.Status, t.Visibility, t.Slug, t.MaxParticipants, t.OwnerUserID, t.RegistrationMode, t.Latitude, t.Longitude)
	return err
}

func (r *TournamentRepository) Update(ctx context.Context, t *entity.Tournament) error {
	_, err := r.db.Exec(ctx, `
        UPDATE tournaments
        SET title=$2, discipline=$3, description=$4, rules=$5, location=$6, max_teams=$7, format=$8, group_count=$9, registration_deadline=$10, start_at=$11, visibility=$12, registration_mode=$13, max_participants=$14, latitude=$15, longitude=$16, updated_at=now()
        WHERE id=$1 AND deleted_at IS NULL
    `, t.ID, t.Title, t.Discipline, t.Description, t.Rules, t.Location, t.MaxTeams, t.Format, t.GroupCount, t.RegistrationDeadline, t.StartAt, t.Visibility, t.RegistrationMode, t.MaxParticipants, t.Latitude, t.Longitude)
	return err
}

func (r *TournamentRepository) SetStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET status=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id, status)
	return err
}

func (r *TournamentRepository) SetWinner(ctx context.Context, tournamentID string, winnerTeamID, winnerParticipantID *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tournaments SET winner_team_id=$2, winner_participant_id=$3, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`,
		tournamentID, winnerTeamID, winnerParticipantID,
	)
	return err
}

func (r *TournamentRepository) UpdateStatus(ctx context.Context, id, status string) error {
	return r.SetStatus(ctx, id, status)
}

func (r *TournamentRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *TournamentRepository) GetByID(ctx context.Context, id string) (*entity.Tournament, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, title, discipline, description, rules, location, max_teams, format, group_count, registration_deadline, start_at, status, visibility, slug, max_participants, owner_user_id, registration_mode, winner_team_id, winner_participant_id, latitude, longitude, created_at, updated_at, deleted_at
        FROM tournaments
        WHERE id=$1 AND deleted_at IS NULL
    `, id)
	return scanTournament(row)
}

func (r *TournamentRepository) GetBySlug(ctx context.Context, slug string) (*entity.Tournament, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, title, discipline, description, rules, location, max_teams, format, group_count, registration_deadline, start_at, status, visibility, slug, max_participants, owner_user_id, registration_mode, winner_team_id, winner_participant_id, latitude, longitude, created_at, updated_at, deleted_at
        FROM tournaments
        WHERE slug=$1 AND deleted_at IS NULL
    `, slug)
	return scanTournament(row)
}

func (r *TournamentRepository) ListPublic(ctx context.Context, limit, offset int, f TournamentFilter) ([]entity.Tournament, error) {
	where, args := publicWhere(f)
	args = append(args, limit, offset)
	q := fmt.Sprintf(`
        SELECT id, title, discipline, description, rules, location, max_teams, format, group_count, registration_deadline, start_at, status, visibility, slug, max_participants, owner_user_id, registration_mode, winner_team_id, winner_participant_id, latitude, longitude, created_at, updated_at, deleted_at
        FROM tournaments
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, where, len(args)-1, len(args))
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTournamentRows(rows)
}

func (r *TournamentRepository) CountPublic(ctx context.Context, f TournamentFilter) (int, error) {
	where, args := publicWhere(f)
	q := fmt.Sprintf(`SELECT COUNT(*) FROM tournaments %s`, where)
	var n int
	err := r.db.QueryRow(ctx, q, args...).Scan(&n)
	return n, err
}

func (r *TournamentRepository) ListAll(ctx context.Context, limit, offset int) ([]entity.Tournament, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, title, discipline, description, rules, location, max_teams, format, group_count, registration_deadline, start_at, status, visibility, slug, max_participants, owner_user_id, registration_mode, winner_team_id, winner_participant_id, latitude, longitude, created_at, updated_at, deleted_at
        FROM tournaments
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTournamentRows(rows)
}

func (r *TournamentRepository) CountAll(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tournaments WHERE deleted_at IS NULL`).Scan(&n)
	return n, err
}

func scanTournament(row interface{ Scan(...interface{}) error }) (*entity.Tournament, error) {
	var t entity.Tournament
	err := row.Scan(&t.ID, &t.Title, &t.Discipline, &t.Description, &t.Rules, &t.Location, &t.MaxTeams, &t.Format, &t.GroupCount, &t.RegistrationDeadline, &t.StartAt, &t.Status, &t.Visibility, &t.Slug, &t.MaxParticipants, &t.OwnerUserID, &t.RegistrationMode, &t.WinnerTeamID, &t.WinnerParticipantID, &t.Latitude, &t.Longitude, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &t, nil
}

func scanTournamentRows(rows interface {
	Next() bool
	Scan(...interface{}) error
	Err() error
}) ([]entity.Tournament, error) {
	result := make([]entity.Tournament, 0)
	for rows.Next() {
		t, err := scanTournament(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *t)
	}
	return result, rows.Err()
}

func (r *TournamentRepository) AddRole(ctx context.Context, role *entity.TournamentUserRole) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO tournament_user_roles (id, tournament_id, user_id, role, assigned_by)
        VALUES ($1,$2,$3,$4,$5)
        ON CONFLICT (tournament_id, user_id, role) DO NOTHING
    `, role.ID, role.TournamentID, role.UserID, role.Role, role.AssignedBy)
	return err
}

func (r *TournamentRepository) RemoveRole(ctx context.Context, tournamentID, userID, role string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM tournament_user_roles WHERE tournament_id=$1 AND user_id=$2 AND role=$3`, tournamentID, userID, role)
	return err
}

func (r *TournamentRepository) HasRole(ctx context.Context, tournamentID, userID, role string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tournament_user_roles WHERE tournament_id=$1 AND user_id=$2 AND role=$3)`, tournamentID, userID, role).Scan(&exists)
	return exists, err
}

func (r *TournamentRepository) ListRoles(ctx context.Context, tournamentID, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT role FROM tournament_user_roles WHERE tournament_id=$1 AND user_id=$2 ORDER BY role`, tournamentID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		result = append(result, role)
	}
	return result, rows.Err()
}

func (r *TournamentRepository) ListUserIDsByRoles(ctx context.Context, tournamentID string, roles []string) ([]string, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT user_id FROM tournament_user_roles WHERE tournament_id=$1 AND role = ANY($2)`, tournamentID, roles)
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
