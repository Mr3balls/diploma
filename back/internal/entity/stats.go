package entity

import "time"

type UserStats struct {
	TournamentsOrganized    int `json:"tournaments_organized"`
	TournamentsParticipated int `json:"tournaments_participated"`
	TournamentsWon          int `json:"tournaments_won"`
	TeamsCount              int `json:"teams_count"`
}

type MyTournamentEntry struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	Format     string     `json:"format"`
	Discipline string     `json:"discipline"`
	StartAt    *time.Time `json:"start_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UserRole   string     `json:"user_role"` // "organizer" | "manager" | "participant"
	IsWinner   bool       `json:"is_winner"`
}
