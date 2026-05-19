package entity

import "time"

type Tournament struct {
	ID                   string     `json:"id"`
	Title                string     `json:"title"`
	Discipline           string     `json:"discipline"`
	Description          *string    `json:"description,omitempty"`
	Rules                *string    `json:"rules,omitempty"`
	Location             *string    `json:"location,omitempty"`
	MaxTeams             int        `json:"max_teams"`
	MaxParticipants      int        `json:"max_participants"`
	Format               string     `json:"format"`
	GroupCount           *int       `json:"group_count,omitempty"`
	RegistrationDeadline *time.Time `json:"registration_deadline,omitempty"`
	StartAt              *time.Time `json:"start_at,omitempty"`
	Status               string     `json:"status"`
	Visibility           string     `json:"visibility"`
	Slug                 *string    `json:"slug,omitempty"`
	OwnerUserID          string     `json:"owner_user_id"`
	RegistrationMode     string     `json:"registration_mode"` // "team" | "individual"
	Latitude            *float64   `json:"latitude,omitempty"`
	Longitude           *float64   `json:"longitude,omitempty"`
	WinnerTeamID        *string    `json:"winner_team_id,omitempty"`
	WinnerParticipantID *string    `json:"winner_participant_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

type TournamentUserRole struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournament_id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"`
	AssignedBy   string    `json:"assigned_by"`
	CreatedAt    time.Time `json:"created_at"`
}
