package entity

import "time"

type Notification struct {
	ID                string     `json:"id"`
	UserID            string     `json:"user_id"`
	Type              string     `json:"type"`
	Title             string     `json:"title"`
	Message           string     `json:"message"`
	PayloadJSON       []byte     `json:"payload_json"`
	ActionPayloadJSON []byte     `json:"action_payload_json"`
	IsRead            bool       `json:"is_read"`
	ActedAt           *time.Time `json:"acted_at,omitempty"`
	ReadAt            *time.Time `json:"read_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty"`
}
