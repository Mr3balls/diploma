package entity

import "time"

type AuthSession struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	UserAgent        *string    `json:"user_agent,omitempty"`
	IPAddress        *string    `json:"ip_address,omitempty"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
