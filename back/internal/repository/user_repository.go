package repository

import (
	"context"
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
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, created_at, updated_at, deleted_at
        FROM users
        WHERE lower(email) = lower($1) AND deleted_at IS NULL
    `
	return scanUser(r.db.QueryRow(ctx, query, email))
}

func (r *UserRepository) GetByNickname(ctx context.Context, nickname string) (*entity.User, error) {
	query := `
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, created_at, updated_at, deleted_at
        FROM users
        WHERE lower(nickname) = lower($1) AND deleted_at IS NULL
    `
	return scanUser(r.db.QueryRow(ctx, query, nickname))
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, created_at, updated_at, deleted_at
        FROM users
        WHERE id = $1 AND deleted_at IS NULL
    `
	return scanUser(r.db.QueryRow(ctx, query, id))
}

func (r *UserRepository) UpdateProfile(ctx context.Context, u *entity.User) error {
	query := `
        UPDATE users
        SET first_name=$2, last_name=$3, nickname=$4, phone=$5, avatar_url=$6, updated_at=now()
        WHERE id=$1 AND deleted_at IS NULL
    `
	_, err := r.db.Exec(ctx, query, u.ID, u.FirstName, u.LastName, u.Nickname, u.Phone, u.AvatarURL)
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
        SELECT id, first_name, last_name, email, phone, nickname, password_hash, avatar_url, is_blocked, created_at, updated_at, deleted_at
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
