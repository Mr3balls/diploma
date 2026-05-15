package entity

import "time"

type Team struct {
	ID                     string     `json:"id"`
	TournamentID           string     `json:"tournament_id"`
	Name                   string     `json:"name"`
	Status                 string     `json:"status"`
	ApprovedByManager      *bool      `json:"approved_by_manager,omitempty"`
	CreatedFromImportRowID *string    `json:"created_from_import_row_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty"`
}

type TeamMember struct {
	ID                 string     `json:"id"`
	TeamID             string     `json:"team_id"`
	UserID             *string    `json:"user_id,omitempty"`
	Nickname           string     `json:"nickname"`
	Email              *string    `json:"email,omitempty"`
	MemberRole         string     `json:"member_role"`
	IsCaptain          bool       `json:"is_captain"`
	IsSubstitute       bool       `json:"is_substitute"`
	ConfirmationStatus string     `json:"confirmation_status"`
	InvitedAt          *time.Time `json:"invited_at,omitempty"`
	RespondedAt        *time.Time `json:"responded_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}
