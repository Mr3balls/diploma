package repository

import (
	"context"
	"time"

	"esports-backend/internal/entity"
)

type ChatRepository struct {
	db Queryer
}

func NewChatRepository(db Queryer) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) Create(ctx context.Context, msg *entity.TournamentMessage) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO tournament_messages (id, tournament_id, user_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, msg.ID, msg.TournamentID, msg.UserID, msg.Content, msg.CreatedAt)
	return err
}

func (r *ChatRepository) GetByID(ctx context.Context, id string) (*entity.TournamentMessage, error) {
	row := r.db.QueryRow(ctx, `
		SELECT m.id, m.tournament_id, m.user_id,
		       COALESCE(u.nickname, u.first_name, u.email) AS user_nickname,
		       m.content, m.created_at
		FROM tournament_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.id = $1
	`, id)
	var m entity.TournamentMessage
	if err := row.Scan(&m.ID, &m.TournamentID, &m.UserID, &m.UserNickname, &m.Content, &m.CreatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

// List returns up to `limit` messages for a tournament, ordered newest-first.
// If `before` is non-zero, only messages older than that timestamp are returned (cursor pagination).
func (r *ChatRepository) List(ctx context.Context, tournamentID string, limit int, before time.Time) ([]*entity.TournamentMessage, error) {
	var (
		rows interface {
			Next() bool
			Scan(...interface{}) error
			Close()
			Err() error
		}
		err error
	)

	const baseQ = `
		SELECT m.id, m.tournament_id, m.user_id,
		       COALESCE(u.nickname, u.first_name, u.email) AS user_nickname,
		       m.content, m.created_at
		FROM tournament_messages m
		JOIN users u ON u.id = m.user_id
		WHERE m.tournament_id = $1
	`

	if before.IsZero() {
		rows, err = r.db.Query(ctx, baseQ+`ORDER BY m.created_at DESC LIMIT $2`, tournamentID, limit)
	} else {
		rows, err = r.db.Query(ctx, baseQ+`AND m.created_at < $2 ORDER BY m.created_at DESC LIMIT $3`, tournamentID, before, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*entity.TournamentMessage, 0)
	for rows.Next() {
		var m entity.TournamentMessage
		if err := rows.Scan(&m.ID, &m.TournamentID, &m.UserID, &m.UserNickname, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}
