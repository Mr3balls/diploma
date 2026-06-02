package entity

import (
	"encoding/json"
	"time"
)

type User struct {
	ID                      string          `json:"id"`
	FirstName               string          `json:"first_name"`
	LastName                string          `json:"last_name"`
	Email                   string          `json:"email"`
	Phone                   string          `json:"phone"`
	Nickname                string          `json:"nickname"`
	PasswordHash            string          `json:"-"`
	AvatarURL               *string         `json:"avatar_url,omitempty"`
	IsBlocked               bool            `json:"is_blocked"`
	Lang                    string          `json:"lang"`
	NotificationPreferences json.RawMessage `json:"notification_preferences,omitempty"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
	DeletedAt               *time.Time      `json:"deleted_at,omitempty"`
	Roles                   []string        `json:"roles,omitempty"`
}
