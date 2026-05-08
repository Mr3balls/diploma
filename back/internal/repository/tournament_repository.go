package repository

import (
	"context"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type TournamentRepository struct {
	db Queryer
}

func NewTournamentRepository(db Queryer) *TournamentRepository {
	return &TournamentRepository{db: db}
}

func (r *TournamentRepository) Create(ctx context.Context, t *entity.Tournament) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO tournaments (id, title, discipline, description, rules, location, max_teams, registration_deadline, start_at, status, visibility, owner_user_id)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
    `, t.ID, t.Title, t.Discipline, t.Description, t.Rules, t.Location, t.MaxTeams, t.RegistrationDeadline, t.StartAt, t.Status, t.Visibility, t.OwnerUserID)
	return err
}

func (r *TournamentRepository) Update(ctx context.Context, t *entity.Tournament) error {
	_, err := r.db.Exec(ctx, `
        UPDATE tournaments
        SET title=$2, discipline=$3, description=$4, rules=$5, location=$6, max_teams=$7, registration_deadline=$8, start_at=$9, visibility=$10, updated_at=now()
        WHERE id=$1 AND deleted_at IS NULL
    `, t.ID, t.Title, t.Discipline, t.Description, t.Rules, t.Location, t.MaxTeams, t.RegistrationDeadline, t.StartAt, t.Visibility)
	return err
}

func (r *TournamentRepository) SetStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET status=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id, status)
	return err
}

func (r *TournamentRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *TournamentRepository) GetByID(ctx context.Context, id string) (*entity.Tournament, error) {
	row := r.db.QueryRow(ctx, `
        SELECT id, title, discipline, description, rules, location, max_teams, registration_deadline, start_at, status, visibility, owner_user_id, created_at, updated_at, deleted_at
        FROM tournaments
        WHERE id=$1 AND deleted_at IS NULL
    `, id)
	var t entity.Tournament
	err := row.Scan(&t.ID, &t.Title, &t.Discipline, &t.Description, &t.Rules, &t.Location, &t.MaxTeams, &t.RegistrationDeadline, &t.StartAt, &t.Status, &t.Visibility, &t.OwnerUserID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &t, nil
}

func (r *TournamentRepository) ListPublic(ctx context.Context, limit, offset int) ([]entity.Tournament, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, title, discipline, description, rules, location, max_teams, registration_deadline, start_at, status, visibility, owner_user_id, created_at, updated_at, deleted_at
        FROM tournaments
        WHERE deleted_at IS NULL AND visibility='public'
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entity.Tournament, 0)
	for rows.Next() {
		var t entity.Tournament
		if err := rows.Scan(&t.ID, &t.Title, &t.Discipline, &t.Description, &t.Rules, &t.Location, &t.MaxTeams, &t.RegistrationDeadline, &t.StartAt, &t.Status, &t.Visibility, &t.OwnerUserID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *TournamentRepository) ListAll(ctx context.Context, limit, offset int) ([]entity.Tournament, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, title, discipline, description, rules, location, max_teams, registration_deadline, start_at, status, visibility, owner_user_id, created_at, updated_at, deleted_at
        FROM tournaments
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entity.Tournament, 0)
	for rows.Next() {
		var t entity.Tournament
		if err := rows.Scan(&t.ID, &t.Title, &t.Discipline, &t.Description, &t.Rules, &t.Location, &t.MaxTeams, &t.RegistrationDeadline, &t.StartAt, &t.Status, &t.Visibility, &t.OwnerUserID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, err
		}
		result = append(result, t)
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
