package entity

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID           string          `json:"id"`
	ActorUserID  *string         `json:"actor_user_id,omitempty"`
	TournamentID *string         `json:"tournament_id,omitempty"`
	EntityType   string          `json:"entity_type"`
	EntityID     string          `json:"entity_id"`
	ActionType   string          `json:"action_type"`
	Description  string          `json:"description"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
	CreatedAt    time.Time       `json:"created_at"`
}
