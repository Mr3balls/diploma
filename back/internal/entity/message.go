package entity

import "time"

type TournamentMessage struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournament_id"`
	UserID       string    `json:"user_id"`
	UserNickname string    `json:"user_nickname"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
}
