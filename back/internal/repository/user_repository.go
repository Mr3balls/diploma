package repository

import (
	"context"
	"encoding/json"
	"errors"

	"esports-backend/internal/entity"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db Queryer
}

func NewUserRepository(db Queryer) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u *entity.User) error {
	query := `
        INSERT INTO users (id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `
	_, err := r.db.Exec(ctx, query, u.ID, u.FirstName, u.LastName, u.Email, u.Phone, u.Nickname, u.PasswordHash, u.AvatarURL, u.IsBlocked)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, lang, notification_preferences, created_at, updated_at, deleted_at
        FROM users
        WHERE lower(email) = lower($1) AND deleted_at IS NULL
    `
	return scanUser(r.db.QueryRow(ctx, query, email))
}

func (r *UserRepository) GetByNickname(ctx context.Context, nickname string) (*entity.User, error) {
	query := `
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, lang, notification_preferences, created_at, updated_at, deleted_at
        FROM users
        WHERE lower(nickname) = lower($1) AND deleted_at IS NULL
    `
	return scanUser(r.db.QueryRow(ctx, query, nickname))
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, lang, notification_preferences, created_at, updated_at, deleted_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `
	return scanUser(r.db.QueryRow(ctx, query, id))
}

func (r *UserRepository) UpdateProfile(ctx context.Context, u *entity.User) error {
	query := `
        UPDATE users
        SET first_name=$2, last_name=$3, nickname=$4, phone=$5, avatar_url=$6, lang=$7, updated_at=now()
        WHERE id=$1 AND deleted_at IS NULL
    `
	_, err := r.db.Exec(ctx, query, u.ID, u.FirstName, u.LastName, u.Nickname, u.Phone, u.AvatarURL, u.Lang)
	return err
}

func (r *UserRepository) GetLangsByIDs(ctx context.Context, ids []string) map[string]string {
	result := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return result
	}
	rows, err := r.db.Query(ctx, `SELECT id, lang FROM users WHERE id = ANY($1) AND deleted_at IS NULL`, ids)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var id, lang string
		if rows.Scan(&id, &lang) == nil && lang != "" {
			result[id] = lang
		}
	}
	return result
}

func (r *UserRepository) GetNotificationPreferences(ctx context.Context, userID string) ([]string, error) {
	var raw []byte
	err := r.db.QueryRow(ctx,
		`SELECT notification_preferences FROM users WHERE id=$1 AND deleted_at IS NULL`, userID,
	).Scan(&raw)
	if err != nil || raw == nil {
		return nil, err
	}
	var p struct {
		Disabled []string `json:"disabled"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, nil
	}
	return p.Disabled, nil
}

func (r *UserRepository) SetNotificationPreferences(ctx context.Context, userID string, disabled []string) error {
	type prefs struct {
		Disabled []string `json:"disabled"`
	}
	data, err := json.Marshal(prefs{Disabled: disabled})
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx,
		`UPDATE users SET notification_preferences=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`,
		userID, data,
	)
	return err
}

func (r *UserRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}

func (r *UserRepository) GetPlatformRoles(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
        SELECT pr.code
        FROM user_platform_roles upr
        JOIN platform_roles pr ON pr.id = upr.role_id
        WHERE upr.user_id = $1
        ORDER BY pr.code
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *UserRepository) AssignPlatformRole(ctx context.Context, userID, roleCode string) error {
	_, err := r.db.Exec(ctx, `
        INSERT INTO user_platform_roles (id, user_id, role_id)
        SELECT gen_random_uuid(), $1, pr.id
        FROM platform_roles pr
        WHERE pr.code = $2
        ON CONFLICT DO NOTHING
    `, userID, roleCode)
	return err
}

func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]entity.User, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, lang, notification_preferences, created_at, updated_at, deleted_at
        FROM users
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]entity.User, 0)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *u)
	}
	return result, rows.Err()
}

func (r *UserRepository) CountUsers(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&n)
	return n, err
}

func (r *UserRepository) SetBlocked(ctx context.Context, userID string, blocked bool) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET is_blocked=$2, updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, userID, blocked)
	return err
}

func (r *UserRepository) GetMyStats(ctx context.Context, userID string) (*entity.UserStats, error) {
	row := r.db.QueryRow(ctx, `
		SELECT
		  (SELECT COUNT(*) FROM tournament_user_roles WHERE user_id=$1 AND role='owner') AS organized,
		  (SELECT COUNT(DISTINCT t_id) FROM (
		    SELECT tournament_id AS t_id FROM participants WHERE user_id=$1
		    UNION
		    SELECT tm_tbl.tournament_id AS t_id
		    FROM team_members tm
		    JOIN teams tm_tbl ON tm_tbl.id = tm.team_id AND tm_tbl.deleted_at IS NULL
		    WHERE tm.user_id=$1
		  ) sub) AS participated,
		  (SELECT COUNT(*) FROM tournaments WHERE deleted_at IS NULL AND (
		    winner_participant_id IN (SELECT id FROM participants WHERE user_id=$1)
		    OR winner_team_id IN (
		      SELECT tm_tbl.id FROM teams tm_tbl
		      JOIN team_members tm ON tm.team_id = tm_tbl.id AND tm.user_id=$1
		      WHERE tm_tbl.deleted_at IS NULL
		    )
		  )) AS won,
		  (SELECT COUNT(DISTINCT team_id) FROM team_members WHERE user_id=$1) AS teams_count
	`, userID)
	var s entity.UserStats
	if err := row.Scan(&s.TournamentsOrganized, &s.TournamentsParticipated, &s.TournamentsWon, &s.TeamsCount); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *UserRepository) GetMyTournaments(ctx context.Context, userID string) ([]entity.MyTournamentEntry, error) {
	rows, err := r.db.Query(ctx, `
		WITH user_tournaments AS (
		  SELECT t.id,
		    CASE WHEN tur.role='owner' THEN 1 WHEN tur.role='manager' THEN 2 ELSE 3 END AS role_rank
		  FROM tournament_user_roles tur
		  JOIN tournaments t ON t.id=tur.tournament_id AND t.deleted_at IS NULL
		  WHERE tur.user_id=$1
		  UNION ALL
		  SELECT t.id, 3
		  FROM participants p
		  JOIN tournaments t ON t.id=p.tournament_id AND t.deleted_at IS NULL
		  WHERE p.user_id=$1
		  UNION ALL
		  SELECT t.id, 3
		  FROM team_members tm
		  JOIN teams tm_tbl ON tm_tbl.id=tm.team_id AND tm_tbl.deleted_at IS NULL
		  JOIN tournaments t ON t.id=tm_tbl.tournament_id AND t.deleted_at IS NULL
		  WHERE tm.user_id=$1
		),
		ranked AS (
		  SELECT id, MIN(role_rank) AS role_rank FROM user_tournaments GROUP BY id
		)
		SELECT
		  t.id, t.title, t.status, t.format,
		  COALESCE(t.discipline,'') AS discipline,
		  t.start_at, t.created_at,
		  CASE r.role_rank WHEN 1 THEN 'organizer' WHEN 2 THEN 'manager' ELSE 'participant' END,
		  COALESCE(
		    t.winner_participant_id IN (SELECT id FROM participants WHERE user_id=$1)
		    OR t.winner_team_id IN (
		      SELECT tm_tbl.id FROM teams tm_tbl
		      JOIN team_members tm ON tm.team_id=tm_tbl.id AND tm.user_id=$1
		      WHERE tm_tbl.deleted_at IS NULL
		    ),
		    false
		  ) AS is_winner
		FROM ranked r
		JOIN tournaments t ON t.id=r.id
		ORDER BY t.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.MyTournamentEntry, 0)
	for rows.Next() {
		var e entity.MyTournamentEntry
		if err := rows.Scan(&e.ID, &e.Title, &e.Status, &e.Format, &e.Discipline, &e.StartAt, &e.CreatedAt, &e.UserRole, &e.IsWinner); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func scanUser(row interface {
	Scan(dest ...interface{}) error
}) (*entity.User, error) {
	var u entity.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Phone,
		&u.Nickname,
		&u.PasswordHash,
		&u.AvatarURL,
		&u.IsBlocked,
		&u.Lang,
		&u.NotificationPreferences,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, err
	}
	return &u, nil
}
