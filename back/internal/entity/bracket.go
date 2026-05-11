package entity

import (
	"encoding/json"
	"time"
)

type Bracket struct {
	ID            string          `json:"id"`
	TournamentID  string          `json:"tournament_id"`
	Format        string          `json:"format"`
	SeedingMethod string          `json:"seeding_method"`
	Status        string          `json:"status"`
	GeneratedBy   string          `json:"generated_by"`
	GeneratedAt   time.Time       `json:"generated_at"`
	MetadataJSON  json.RawMessage `json:"metadata_json"`
}

type Match struct {
	ID                      string     `json:"id"`
	TournamentID            string     `json:"tournament_id"`
	BracketID               string     `json:"bracket_id"`
	BracketSection          string     `json:"bracket_section"` // WB | LB | GF
	RoundNumber             int        `json:"round_number"`
	SlotIndex               int        `json:"slot_index"`
	GlobalNumber            int        `json:"global_number,omitempty"`
	// Team-based tournament fields
	Team1ID                 *string    `json:"team1_id,omitempty"`
	Team2ID                 *string    `json:"team2_id,omitempty"`
	// Challonge-style participant fields
	Participant1ID          *string    `json:"participant1_id,omitempty"`
	Participant2ID          *string    `json:"participant2_id,omitempty"`
	WinnerParticipantID     *string    `json:"winner_participant_id,omitempty"`
	ScheduledAt             *time.Time `json:"scheduled_at,omitempty"`
	LocationOrServer        *string    `json:"location_or_server,omitempty"`
	Status                  string     `json:"status"`
	Team1ConfirmationStatus string     `json:"team1_confirmation_status"`
	Team2ConfirmationStatus string     `json:"team2_confirmation_status"`
	WinnerTeamID            *string    `json:"winner_team_id,omitempty"`
	ScoreText               *string    `json:"score_text,omitempty"`
	ManagerComment          *string    `json:"manager_comment,omitempty"`
	NextMatchID             *string    `json:"next_match_id,omitempty"`
	SourceMatch1ID          *string    `json:"source_match1_id,omitempty"`
	SourceMatch2ID          *string    `json:"source_match2_id,omitempty"`
	LoserNextMatchID        *string    `json:"loser_next_match_id,omitempty"`
	LoserNextSlot           int        `json:"loser_next_slot,omitempty"`
	IsBye                   bool       `json:"is_bye"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
	DeletedAt               *time.Time `json:"deleted_at,omitempty"`
}
