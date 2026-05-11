package entity

import (
	"encoding/json"
	"time"
)

// Participant is a single player in a Challonge-style tournament.
// UserID is nil when the participant was added manually without linking to a user account.
type Participant struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournament_id"`
	UserID       *string   `json:"user_id,omitempty"`
	Name         string    `json:"name"`
	Seed         int       `json:"seed"`
	Status       string    `json:"status"` // active | eliminated | champion
	FinalRank    *int      `json:"final_rank,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// TournamentMember maps a user to a Challonge-style tournament role.
// Separate from TournamentUserRole (owner/manager) to avoid schema collisions.
type TournamentMember struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournament_id"`
	UserID       string    `json:"user_id"`
	Role         string    `json:"role"` // organizer | co_organizer | participant | viewer
	JoinedAt     time.Time `json:"joined_at"`
}

// MatchLogEntry records a single action on a match for the audit trail.
type MatchLogEntry struct {
	ID           string          `json:"id"`
	TournamentID string          `json:"tournament_id"`
	MatchID      string          `json:"match_id"`
	Action       string          `json:"action"`
	ActorID      string          `json:"actor_id"`
	Detail       json.RawMessage `json:"detail,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// ResultReport is a match result submitted by a participant awaiting organizer approval.
type ResultReport struct {
	ID           string    `json:"id"`
	MatchID      string    `json:"match_id"`
	ReportedByID string    `json:"reported_by_id"`
	WinnerID     string    `json:"winner_id"` // participant ID
	Score1       int       `json:"score1"`
	Score2       int       `json:"score2"`
	Status       string    `json:"status"` // pending | approved | rejected
	CreatedAt    time.Time `json:"created_at"`
}

// CoOrganizerInvite is a token-based invitation to become a co-organizer.
// Token is excluded from JSON responses for security.
type CoOrganizerInvite struct {
	ID           string    `json:"id"`
	TournamentID string    `json:"tournament_id"`
	InviteeID    string    `json:"invitee_id"`
	InvitedByID  string    `json:"invited_by_id"`
	Token        string    `json:"-"`
	Status       string    `json:"status"` // pending | accepted | rejected | expired
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}
